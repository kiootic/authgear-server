import { useCallback, useMemo, useState } from "react";
import deepEqual from "deep-equal";
import { useAppConfigQuery } from "../graphql/portal/query/appConfigQuery";
import { useUpdateAppConfigMutation } from "../graphql/portal/mutations/updateAppConfigMutation";
import { PortalAPIAppConfig } from "../types";

export interface AppConfigFormModel<State> {
  isLoading: boolean;
  isUpdating: boolean;
  isDirty: boolean;
  loadError: unknown;
  updateError: unknown;
  state: State;
  setState: (fn: (state: State) => State) => void;
  reload: () => void;
  reset: () => void;
  save: () => void;
}

export type StateConstructor<State> = (config: PortalAPIAppConfig) => State;
export type ConfigConstructor<State> = (
  config: PortalAPIAppConfig,
  initialState: State,
  currentState: State,
  effectiveConfig: PortalAPIAppConfig
) => PortalAPIAppConfig;

export function useAppConfigForm<State>(
  appID: string,
  constructState: StateConstructor<State>,
  constructConfig: ConfigConstructor<State>
): AppConfigFormModel<State> {
  const {
    loading: isLoading,
    error: loadError,
    effectiveAppConfig,
    rawAppConfig: rawConfig,
    refetch: reload,
  } = useAppConfigQuery(appID);
  const {
    loading: isUpdating,
    error: updateError,
    updateAppConfig: updateConfig,
    resetError,
  } = useUpdateAppConfigMutation(appID);

  const effectiveConfig = useMemo(() => effectiveAppConfig ?? { id: appID }, [
    effectiveAppConfig,
    appID,
  ]);

  const initialState = useMemo(() => constructState(effectiveConfig), [
    effectiveConfig,
    constructState,
  ]);
  const [currentState, setCurrentState] = useState<State | null>(null);

  const isDirty = useMemo(() => {
    if (!rawConfig || !currentState) {
      return false;
    }
    return !deepEqual(
      constructConfig(rawConfig, initialState, initialState, effectiveConfig),
      constructConfig(rawConfig, initialState, currentState, effectiveConfig),
      { strict: true }
    );
  }, [constructConfig, rawConfig, initialState, currentState, effectiveConfig]);

  const reset = useCallback(() => {
    if (isUpdating) {
      return;
    }
    resetError();
    setCurrentState(null);
  }, [isUpdating, resetError]);

  const save = useCallback(() => {
    if (!rawConfig || !initialState || !currentState) {
      return;
    } else if (!isDirty || isUpdating) {
      return;
    }

    const newConfig = constructConfig(
      rawConfig,
      initialState,
      currentState,
      effectiveConfig
    );
    updateConfig(newConfig)
      .then(() => setCurrentState(null))
      .catch(() => {});
  }, [
    isDirty,
    isUpdating,
    constructConfig,
    rawConfig,
    effectiveConfig,
    initialState,
    currentState,
    updateConfig,
  ]);

  const state = currentState ?? initialState;
  const setState = useCallback(
    (fn: (state: State) => State) => {
      setCurrentState(fn(state));
    },
    [state]
  );

  return {
    isLoading,
    isUpdating,
    isDirty,
    loadError,
    updateError,
    state,
    setState,
    reload,
    reset,
    save,
  };
}