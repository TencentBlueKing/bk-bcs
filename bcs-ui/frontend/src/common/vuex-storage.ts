import { Store, MutationPayload } from 'vuex';
import { merge, get, set } from 'lodash';

interface Storage {
  getItem: (key: string) => any;
  setItem: (key: string, value: any) => void;
  removeItem: (key: string) => void;
}
interface Effect {
  type: string
  effect: (state, store) => void
}

interface Options<State> {
  key: string;
  overwrite: boolean;
  storage: Storage;
  paths: string[];
  reducer: (state: State, paths: string[]) => object;
  subscriber: (
    store: Store<State>
  ) => (handler: (mutation: any, state: State) => void) => void;
  getState: (key: string, storage: Storage) => any;
  setState: (key: string, state: any, storage: Storage) => void;
  assertStorage?: (storage: Storage) => void | Error;
  mutationEffect?: Effect[]
}

function reducer(state, paths) {
  return Array.isArray(paths)
    ? paths.reduce((substate, path) => set(substate, path, get(state, path)), {})
    : state;
}

function subscriber(store) {
  return function (handler) {
    return store.subscribe(handler);
  };
}

function getState(key = 'vuex', storage: Storage = window.localStorage) {
  const value = storage.getItem(key);

  try {
    return !!value ? JSON.parse(value) : value;
  } catch (err) {}

  return undefined;
}

function setState(key = 'vuex', state, storage) {
  return storage.setItem(key, JSON.stringify(state));
}

function assertStorage(storage: Storage = window.localStorage) {
  storage.setItem('@@', 1);
  storage.removeItem('@@');
}

export default <State>(opt: Partial<Options<State>> = {}) => {
  const options: Options<State> = merge({
    key: 'vuex',
    overwrite: false,
    storage: window.localStorage,
    paths: [],
    reducer,
    subscriber,
    getState,
    setState,
    assertStorage,
  }, opt);

  assertStorage(options.storage);
  const savedState = options.getState(options.key, options.storage);
  const effectMap = (options.mutationEffect || []).reduce<Record<string, Effect['effect']>>((pre, curent) => {
    if (pre[curent.type]) {
      console.warn(`repeat type ${curent.type}`);
    }
    pre[curent.type] = curent.effect;
    return pre;
  }, {});

  return (store: Store<State>) => {
    if (typeof savedState === 'object' && savedState !== null) {
      store.replaceState(options.overwrite
        ? savedState
        : merge(
          store.state,
          savedState,
        ));
    }

    options.subscriber(store)((mutation: MutationPayload, state) => {
      effectMap[mutation.type]?.(state, store);
      options.setState(
        options.key,
        options.reducer(state, options.paths),
        options.storage,
      );
    });
  };
};
