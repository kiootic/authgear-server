package pq

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	sq "github.com/lann/squirrel"
	"github.com/lib/pq"

	"github.com/oursky/ourd/oddb"
)

// This file implements Record related operations of the
// oddb/pq implementation.

// Different data types that can be saved in and loaded from postgreSQL
// NOTE(limouren): varchar is missing because text can replace them,
// see the docs here: http://www.postgresql.org/docs/9.4/static/datatype-character.html
const (
	TypeString    = "text"
	TypeNumber    = "double precision"
	TypeBoolean   = "boolean"
	TypeJSON      = "jsonb"
	TypeTimestamp = "timestamp without time zone"
)

type nullJSON struct {
	JSON  interface{}
	Valid bool
}

func (nj *nullJSON) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if value == nil || !ok {
		nj.JSON = nil
		nj.Valid = false
		return nil
	}

	err := json.Unmarshal(data, &nj.JSON)
	nj.Valid = err == nil

	return err
}

type referenceValue oddb.Reference

func (ref referenceValue) Value() (driver.Value, error) {
	return ref.ID.Key, nil
}

type jsonSliceValue []interface{}

func (s jsonSliceValue) Value() (driver.Value, error) {
	return json.Marshal([]interface{}(s))
}

type jsonMapValue map[string]interface{}

func (m jsonMapValue) Value() (driver.Value, error) {
	return json.Marshal(map[string]interface{}(m))
}

type aclValue oddb.RecordACL

func (acl aclValue) Value() (driver.Value, error) {
	return json.Marshal(acl)
}

func (db *database) Get(id oddb.RecordID, record *oddb.Record) error {
	typemap, err := db.remoteColumnTypes(id.Type)
	if err != nil {
		return err
	}

	if len(typemap) == 0 { // record type has not been created
		return oddb.ErrRecordNotFound
	}

	sql, args, err := db.selectQuery(id.Type, typemap).ToSql()
	if err != nil {
		panic(err)
	}

	log.WithFields(log.Fields{
		"sql":  sql,
		"args": args,
	}).Debugln("Getting record")

	row := db.Db.QueryRowx(sql, args...)
	if err := newRecordScanner(id.Type, typemap, row).Scan(record); err != nil {
		return err
	}
	return nil
}

// Save attempts to do a upsert
func (db *database) Save(record *oddb.Record) error {
	if record.ID.Key == "" {
		return errors.New("db.save: got empty record id")
	}
	if record.ID.Type == "" {
		return fmt.Errorf("db.save %s: got empty record type", record.ID.Key)
	}
	if record.OwnerID == "" {
		return fmt.Errorf("db.save %s: got empty OwnerID", record.ID.Key)
	}

	sql, args := upsertQuery(db.tableName(record.ID.Type), map[string]interface{}{
		"_id":          record.ID.Key,
		"_database_id": db.userID,
	}, convert(record), []string{"_owner_id"})

	_, err := db.Db.Exec(sql, args...)
	if err != nil {
		log.WithFields(log.Fields{
			"sql":  sql,
			"args": args,
			"err":  err,
		}).Debugln("Failed to save record")

		return err
	}

	record.DatabaseID = db.userID
	return nil
}

func convert(r *oddb.Record) map[string]interface{} {
	m := map[string]interface{}{}
	for key, rawValue := range r.Data {
		switch value := rawValue.(type) {
		default:
			m[key] = rawValue
		case []interface{}:
			m[key] = jsonSliceValue(value)
		case map[string]interface{}:
			m[key] = jsonMapValue(value)
		case oddb.Reference:
			m[key] = referenceValue(value)
		}
	}
	m["_owner_id"] = r.OwnerID
	m["_access"] = aclValue(r.ACL)
	return m
}

func (db *database) Delete(id oddb.RecordID) error {
	sql, args, err := psql.Delete(db.tableName(id.Type)).
		Where("_id = ? AND _database_id = ?", id.Key, db.userID).
		ToSql()
	if err != nil {
		panic(err)
	}

	log.WithFields(log.Fields{
		"sql":  sql,
		"args": args,
	}).Debug("Executing SQL")

	result, err := db.Db.Exec(sql, args...)
	if isUndefinedTable(err) {
		return oddb.ErrRecordNotFound
	} else if err != nil {
		log.WithFields(log.Fields{
			"id":   id,
			"sql":  sql,
			"args": args,
			"err":  err,
		}).Errorf("Failed to execute delete record statement")
		return fmt.Errorf("delete %s: failed to delete record", id)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.WithFields(log.Fields{
			"id":   id,
			"sql":  sql,
			"args": args,
			"err":  err,
		}).Errorln("Failed to fetch rowsAffected")
		return fmt.Errorf("delete %s: failed to retrieve deletion status", id)
	}

	if rowsAffected == 0 {
		return oddb.ErrRecordNotFound
	} else if rowsAffected > 1 {
		log.WithFields(log.Fields{
			"id":           id,
			"sql":          sql,
			"args":         args,
			"rowsAffected": rowsAffected,
			"err":          err,
		}).Errorln("Unexpected rows deleted")
		return fmt.Errorf("delete %s: got %v rows deleted, want 1", id, rowsAffected)
	}

	return err
}

type compoundPredicateSqlizer struct {
	oddb.Predicate
}

type comparisonPredicateSqlizer struct {
	oddb.Predicate
}

func newPredicateSqlizer(predicate oddb.Predicate) sq.Sqlizer {
	if predicate.Operator.IsCompound() {
		return &compoundPredicateSqlizer{predicate}
	} else {
		return &comparisonPredicateSqlizer{predicate}
	}
}

func (p *compoundPredicateSqlizer) ToSql() (sql string, args []interface{}, err error) {
	switch p.Operator {
	default:
		err = fmt.Errorf("Compound operator `%v` is not supported.", p.Operator)
		return
	case oddb.And:
		and := make(sq.And, len(p.Children))
		for i, child := range p.Children {
			and[i] = newPredicateSqlizer(child.(oddb.Predicate))
		}
		return and.ToSql()
	case oddb.Or:
		or := make(sq.Or, len(p.Children))
		for i, child := range p.Children {
			or[i] = newPredicateSqlizer(child.(oddb.Predicate))
		}
		return or.ToSql()
	case oddb.Not:
		pred := p.Children[0].(oddb.Predicate)
		sql, args, err = newPredicateSqlizer(pred).ToSql()
		sql = fmt.Sprintf("NOT (%s)", sql)
		return
	}
}

func toSqlOperand(expr oddb.Expression) (sql string, arg interface{}) {
	if expr.Type == oddb.KeyPath {
		sql = strconv.Quote(expr.Value.(string))
		arg = nil
	} else {
		sql = "?"
		switch svalue := expr.Value.(type) {
		case oddb.Reference:
			arg = svalue.ID.Key
		default:
			arg = svalue
		}
	}
	return
}

func (p *comparisonPredicateSqlizer) ToSql() (sql string, args []interface{}, err error) {
	args = make([]interface{}, 0, 1)
	if p.Operator.IsBinary() {
		var buffer bytes.Buffer
		lhs := p.Children[0].(oddb.Expression)
		rhs := p.Children[1].(oddb.Expression)

		sqlOperand, arg := toSqlOperand(lhs)
		buffer.WriteString(sqlOperand)
		if arg != nil {
			args = append(args, arg)
		}

		switch p.Operator {
		default:
			err = fmt.Errorf("Comparison operator `%v` is not supported.", p.Operator)
			return
		case oddb.Equal:
			buffer.WriteString(`=`)
		case oddb.GreaterThan:
			buffer.WriteString(`>`)
		case oddb.LessThan:
			buffer.WriteString(`<`)
		case oddb.GreaterThanOrEqual:
			buffer.WriteString(`>=`)
		case oddb.LessThanOrEqual:
			buffer.WriteString(`<=`)
		case oddb.NotEqual:
			buffer.WriteString(`<>`)
		}

		sqlOperand, arg = toSqlOperand(rhs)
		buffer.WriteString(sqlOperand)
		if arg != nil {
			args = append(args, arg)
		}

		sql = buffer.String()
		return
	} else {
		err = fmt.Errorf("Comparison operator `%v` is not supported.", p.Operator)
		return
	}
}

func (db *database) Query(query *oddb.Query) (*oddb.Rows, error) {
	if query.Type == "" {
		return nil, errors.New("got empty query type")
	}

	typemap, err := db.remoteColumnTypes(query.Type)
	if err != nil {
		return nil, err
	}

	if len(typemap) == 0 { // record type has not been created
		return oddb.EmptyRows, nil
	}

	q := db.selectQuery(query.Type, typemap)

	if p := query.Predicate; p != nil {
		q = q.Where(newPredicateSqlizer(*p))
	}

	for _, sort := range query.Sorts {
		switch sort.Order {
		default:
			return nil, fmt.Errorf("unknown sort order = %v", sort.Order)
		// NOTE(limouren): better to verify KeyPath as well
		case oddb.Asc:
			q = q.OrderBy(`"` + sort.KeyPath + `"` + " ASC")
		case oddb.Desc:
			q = q.OrderBy(`"` + sort.KeyPath + `"` + " DESC")
		}
	}

	if query.ReadableBy != "" {
		// FIXME: Serialize the json instead of building manually
		q = q.Where(
			`(_access @> '[{"user_id":"`+query.ReadableBy+`"}]' OR `+
				`_access IS NULL OR `+
				`_owner_id = ?)`, query.ReadableBy)
	}

	sql, args, err := q.ToSql()
	if err != nil {
		panic(err)
	}

	log.WithFields(log.Fields{
		"sql":  sql,
		"args": args,
	}).Debugln("Querying record")

	rows, err := db.Db.Queryx(sql, args...)
	return newRows(query.Type, typemap, rows, err)
}

// columnsScanner wraps over sqlx.Rows and sqlx.Row to provide
// a consistent interface for column scanning.
type columnsScanner interface {
	Columns() ([]string, error)
	Scan(dest ...interface{}) error
}

type recordScanner struct {
	recordType string
	typemap    oddb.RecordSchema
	cs         columnsScanner
	columns    []string
	err        error
}

func newRecordScanner(recordType string, typemap oddb.RecordSchema, cs columnsScanner) *recordScanner {
	columns, err := cs.Columns()
	return &recordScanner{recordType, typemap, cs, columns, err}
}

func (rs *recordScanner) Scan(record *oddb.Record) error {
	if rs.err != nil {
		return rs.err
	}

	values := make([]interface{}, 0, len(rs.columns))
	for _, column := range rs.columns {
		schema, ok := rs.typemap[column]
		if !ok {
			return fmt.Errorf("received unknown column = %s", column)
		}
		switch schema.Type {
		default:
			return fmt.Errorf("received unknown data type = %v for column = %s", schema.Type, column)
		case oddb.TypeNumber:
			var number sql.NullFloat64
			values = append(values, &number)
		case oddb.TypeString, oddb.TypeReference, oddb.TypeACL:
			var str sql.NullString
			values = append(values, &str)
		case oddb.TypeDateTime:
			var ts pq.NullTime
			values = append(values, &ts)
		case oddb.TypeBoolean:
			var boolean sql.NullBool
			values = append(values, &boolean)
		case oddb.TypeJSON:
			var j nullJSON
			values = append(values, &j)
		}
	}

	if err := rs.cs.Scan(values...); err != nil {
		rs.err = err
		return err
	}

	record.ID.Type = rs.recordType
	record.Data = map[string]interface{}{}

	for i, column := range rs.columns {
		value := values[i]
		switch svalue := value.(type) {
		default:
			return fmt.Errorf("received unexpected scanned type = %T for column = %s", value, column)
		case *sql.NullFloat64:
			if svalue.Valid {
				record.Set(column, svalue.Float64)
			}
		case *sql.NullString:
			if svalue.Valid {
				schema := rs.typemap[column]
				if schema.Type == oddb.TypeReference {
					record.Set(column, oddb.NewReference(schema.ReferenceType, svalue.String))
				} else if schema.Type == oddb.TypeACL {
					acl := oddb.RecordACL{}
					json.Unmarshal([]byte(svalue.String), &acl)
					record.Set(column, acl)
				} else {
					record.Set(column, svalue.String)
				}
			}
		case *pq.NullTime:
			if svalue.Valid {
				record.Set(column, svalue.Time)
			}
		case *sql.NullBool:
			if svalue.Valid {
				record.Set(column, svalue.Bool)
			}
		case *nullJSON:
			if svalue.Valid {
				record.Set(column, svalue.JSON)
			}
		}
	}

	return nil
}

type rowsIter struct {
	rows *sqlx.Rows
	rs   *recordScanner
}

func (rowsi rowsIter) Close() error {
	return rowsi.rows.Close()
}

func (rowsi rowsIter) Next(record *oddb.Record) error {
	if rowsi.rows.Next() {
		return rowsi.rs.Scan(record)
	} else if rowsi.rows.Err() != nil {
		return rowsi.rows.Err()
	} else {
		return io.EOF
	}
}

func newRows(recordType string, typemap oddb.RecordSchema, rows *sqlx.Rows, err error) (*oddb.Rows, error) {
	if err != nil {
		return nil, err
	}
	rs := newRecordScanner(recordType, typemap, rows)
	return oddb.NewRows(rowsIter{rows, rs}), nil
}

func (db *database) selectQuery(recordType string, typemap oddb.RecordSchema) sq.SelectBuilder {
	columns := make([]string, 0, len(typemap))
	for column := range typemap {
		columns = append(columns, column)
	}

	q := psql.Select()
	for column := range typemap {
		q = q.Column(`"` + column + `"`)
	}

	q = q.From(db.tableName(recordType)).
		Where("_database_id = ?", db.userID)

	return q
}

// SELECT column_name, data_type FROM information_schema.columns
// WHERE table_schema = 'app__' AND table_name = 'note';
// Example for fk
// SELECT
//     tc.table_name, kcu.column_name,
//     ccu.table_name AS foreign_table_name,
//     ccu.column_name AS foreign_column_name
// FROM
//     information_schema.table_constraints AS tc
//     JOIN information_schema.key_column_usage
//         AS kcu ON tc.constraint_name = kcu.constraint_name
//     JOIN information_schema.constraint_column_usage
//         AS ccu ON ccu.constraint_name = tc.constraint_name
// WHERE constraint_type = 'FOREIGN KEY'
// AND tc.table_schema = 'app__'
// AND tc.table_name = 'note';
func (db *database) remoteColumnTypes(recordType string) (oddb.RecordSchema, error) {
	sql, args, err := psql.Select("column_name", "data_type").
		From("information_schema.columns").
		Where("table_schema = ? AND table_name = ?", db.schemaName(), recordType).ToSql()
	if err != nil {
		panic(err)
	}

	log.WithFields(log.Fields{
		"sql":  sql,
		"args": args,
	}).Debugln("Querying columns schema")

	rows, err := db.Db.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	typemap := oddb.RecordSchema{}

	var columnName, pqType string
	for rows.Next() {
		if err := rows.Scan(&columnName, &pqType); err != nil {
			return nil, err
		}

		schema := oddb.FieldType{}
		switch pqType {
		default:
			return nil, fmt.Errorf("received unknown data type = %s for column = %s", pqType, columnName)
		case TypeString:
			schema.Type = oddb.TypeString
		case TypeNumber:
			schema.Type = oddb.TypeNumber
		case TypeTimestamp:
			schema.Type = oddb.TypeDateTime
		case TypeJSON:
			if columnName == "_access" {
				schema.Type = oddb.TypeACL
			} else {
				schema.Type = oddb.TypeJSON
			}
		}

		typemap[columnName] = schema
	}

	// FOREIGN KEY, assumeing we can only reference _id i.e. "ccu.column_name" = _id
	sql, args, err = psql.Select("kcu.column_name", "ccu.table_name").
		From("information_schema.table_constraints AS tc").
		Join("information_schema.key_column_usage AS kcu ON tc.constraint_name = kcu.constraint_name").
		Join("information_schema.constraint_column_usage AS ccu ON ccu.constraint_name = tc.constraint_name").
		Where("constraint_type = 'FOREIGN KEY' AND tc.table_schema = ? AND tc.table_name = ?", db.schemaName(), recordType).ToSql()
	if err != nil {
		panic(err)
	}
	log.WithFields(log.Fields{
		"sql":  sql,
		"args": args,
	}).Debugln("Querying columns referencing schema")
	refs, err := db.Db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	var cName string
	for refs.Next() {
		s := oddb.FieldType{
			Type: oddb.TypeReference,
		}
		if err := refs.Scan(&cName, &s.ReferenceType); err != nil {
			log.Debugf("err %v", err)
			return nil, err
		}
		typemap[cName] = s
		log.Debugln(cName)
		log.Debugln(typemap[cName])
	}
	return typemap, nil
}

func (db *database) Extend(recordType string, recordSchema oddb.RecordSchema) error {
	remoteRecordSchema, err := db.remoteColumnTypes(recordType)
	if err != nil {
		return err
	}

	if len(remoteRecordSchema) == 0 {
		if err := db.createTable(recordType); err != nil {
			return fmt.Errorf("failed to create table: %s", err)
		}
	}
	updatingSchema := oddb.RecordSchema{}
	for key, schema := range recordSchema {
		remoteSchema, ok := remoteRecordSchema[key]
		if !ok {
			updatingSchema[key] = schema
		} else if remoteSchema.Type != schema.Type {
			return fmt.Errorf("conflicting schema %s => %s", remoteSchema, schema)
		}

		// same data type, do nothing
	}

	if len(updatingSchema) > 0 {
		stmt := db.addColumnStmt(recordType, updatingSchema)

		log.WithField("stmt", stmt).Debugln("Adding columns to table")
		if _, err := db.Db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to alter table: %s", err)
		}
	}

	return nil
}

func (db *database) createTable(recordType string) (err error) {
	tablename := db.tableName(recordType)

	stmt := createTableStmt(tablename)
	log.WithField("stmt", stmt).Debugln("Creating table")
	_, err = db.Db.Exec(stmt)
	if err != nil {
		return err
	}

	const CreateTriggerStmtFmt = `CREATE TRIGGER trigger_notify_record_change
	AFTER INSERT OR UPDATE OR DELETE ON %s FOR EACH ROW
	EXECUTE PROCEDURE public.notify_record_change();
`
	stmt = fmt.Sprintf(CreateTriggerStmtFmt, tablename)
	log.WithField("stmt", stmt).Debugln("Creating trigger")
	_, err = db.Db.Exec(stmt)

	return err
}

func createTableStmt(tableName string) string {
	buf := bytes.Buffer{}
	buf.Write([]byte("CREATE TABLE "))
	buf.WriteString(tableName)
	buf.Write([]byte("(_id text, _database_id text, _owner_id text, _access jsonb,"))
	buf.Write([]byte("PRIMARY KEY(_id, _database_id, _owner_id), UNIQUE (_id));"))

	return buf.String()
}

// ALTER TABLE app__.note add collection text;
// ALTER TABLE app__.note
// ADD CONSTRAINT fk_note_collection_collection
// FOREIGN KEY (collection)
// REFERENCES app__.collection(_id);
func (db *database) addColumnStmt(recordType string, recordSchema oddb.RecordSchema) string {
	buf := bytes.Buffer{}
	buf.Write([]byte("ALTER TABLE "))
	buf.WriteString(db.tableName(recordType))
	buf.WriteByte(' ')
	for column, schema := range recordSchema {
		buf.Write([]byte("ADD "))
		buf.WriteByte('"')
		buf.WriteString(column)
		buf.WriteByte('"')
		buf.WriteByte(' ')
		buf.WriteString(pqDataType(schema.Type))
		buf.WriteByte(',')
		switch schema.Type {
		case oddb.TypeAsset:
			db.writeForeignKeyConstraint(&buf, column, "_asset", "id")
		case oddb.TypeReference:
			db.writeForeignKeyConstraint(&buf, column, schema.ReferenceType, "_id")
		}
	}

	// remote the last ','
	buf.Truncate(buf.Len() - 1)

	return buf.String()
}

func (db *database) writeForeignKeyConstraint(buf *bytes.Buffer, localCol, referent, remoteCol string) {
	buf.Write([]byte("ADD CONSTRAINT fk_"))
	buf.WriteString(localCol)
	buf.Write([]byte("_"))
	buf.WriteString(referent)
	buf.Write([]byte("_"))
	buf.WriteString(remoteCol)
	buf.Write([]byte(` FOREIGN KEY ("`))
	buf.WriteString(localCol)
	buf.Write([]byte(`") REFERENCES `))
	buf.WriteString(db.tableName(referent))
	buf.Write([]byte(` ("`))
	buf.WriteString(remoteCol)
	buf.Write([]byte(`"),`))
}

func pqDataType(dataType oddb.DataType) string {
	switch dataType {
	default:
		panic(fmt.Sprintf("Unsupported dataType = %s", dataType))
	case oddb.TypeString, oddb.TypeAsset, oddb.TypeReference:
		return TypeString
	case oddb.TypeNumber:
		return TypeNumber
	case oddb.TypeDateTime:
		return TypeTimestamp
	case oddb.TypeBoolean:
		return TypeBoolean
	case oddb.TypeJSON:
		return TypeJSON
	}
}
