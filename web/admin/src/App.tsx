import { Suspense } from 'react'
import { Spin } from 'antd'
import { RouterProvider } from 'react-router-dom'
import { router } from './routes'

export default function App() {
  return <Suspense fallback={<Spin fullscreen />}><RouterProvider router={router} /></Suspense>
}
