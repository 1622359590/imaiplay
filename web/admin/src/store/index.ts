import { configureStore } from '@reduxjs/toolkit'
import { AUTH_SESSION_EXPIRED_EVENT } from '../api/authSession'
import userReducer, { clearSession } from './userSlice'

export const store = configureStore({
  reducer: {
    user: userReducer,
  },
})

window.addEventListener(AUTH_SESSION_EXPIRED_EVENT, () => {
  store.dispatch(clearSession())
})

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch
