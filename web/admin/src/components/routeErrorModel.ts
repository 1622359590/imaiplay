export interface RouteErrorPresentation {
  title: string
  description: string
}

export function routeErrorPresentation(_error: unknown): RouteErrorPresentation {
  return {
    title: '页面资源加载失败',
    description: '系统可能刚刚完成更新，或当前网络暂时不可用。请刷新页面后重试。',
  }
}
