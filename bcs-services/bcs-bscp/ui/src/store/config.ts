interface State {
    appData: AppData,
    currentVersion: VerionConfig
}

interface AppData {
    id: number;
    spec: {
        name: string;
    }
}

interface VerionConfig {
    attachment: object;
    id: number;
    revision: object;
    spec: object;
}

export default {
    namespaced: true,
    state: {
        appData: {},
        currentVersion: {}
    },
    getters: {
        appName (state: State) {
            return state.appData.spec?.name || ''
        }
    },
    mutations: {
        setAppData (state: State, payload: AppData) {
            state.appData = payload
        },
        setCurrentVersion (state: State, payload: VerionConfig) {
            state.currentVersion = payload
        }
    }
}
