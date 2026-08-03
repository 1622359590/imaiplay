import { createSlice, type PayloadAction } from '@reduxjs/toolkit'
import type { AuthUser } from '../api/auth'
import { readAdminAccessToken } from '../api/authSession'

interface UserState {
  token: string | null
  profile: AuthUser | null
}

const initialState: UserState = {
  token: readAdminAccessToken(),
  profile: null,
}

const userSlice = createSlice({
  name: 'user',
  initialState,
  reducers: {
    setSession(state, action: PayloadAction<{ token: string; user?: AuthUser }>) {
      state.token = action.payload.token
      state.profile = action.payload.user || null
    },
    clearSession(state) {
      state.token = null
      state.profile = null
    },
  },
})

export const { setSession, clearSession } = userSlice.actions
export default userSlice.reducer
