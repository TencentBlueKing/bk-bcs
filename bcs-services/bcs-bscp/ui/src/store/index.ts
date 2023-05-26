import { createPinia } from 'pinia'
import { cloneDeep } from 'lodash'

const pinia = createPinia()

pinia.use(({ store }) => {
  const initialState = cloneDeep(store.$state)
  store.$reset = () => store.$patch(cloneDeep(initialState))
})

export { pinia }
