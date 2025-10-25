/**
 * Dashboard API 服务
 */
import request from './request'
import type { DashboardData } from '@payment/shared/types'

/**
 * 获取Dashboard数据
 * @param timePeriod 时间周期: today | 7days | 30days
 */
export const getDashboardData = async (
  timePeriod: 'today' | '7days' | '30days' = 'today',
): Promise<DashboardData> => {
  // 响应拦截器已解包，直接返回数据
  const response = await request.get<DashboardData>('/dashboard', {
    params: { period: timePeriod },
  })
  return response
}

/**
 * 获取Dashboard统计数据
 */
export const getDashboardStats = async () => {
  const { data } = await request.get('/dashboard/stats')
  return data
}

/**
 * 获取交易趋势数据
 */
export const getTrendData = async (period: 'today' | '7days' | '30days' = 'today') => {
  const { data } = await request.get('/dashboard/trend', {
    params: { period },
  })
  return data
}

/**
 * 获取渠道分布数据
 */
export const getChannelDistribution = async () => {
  const { data } = await request.get('/dashboard/channel-distribution')
  return data
}

/**
 * 获取商户排行
 */
export const getMerchantRanks = async (limit = 5) => {
  const { data } = await request.get('/dashboard/merchant-ranks', {
    params: { limit },
  })
  return data
}

/**
 * 获取近期活动
 */
export const getRecentActivities = async (limit = 5) => {
  const { data } = await request.get('/dashboard/recent-activities', {
    params: { limit },
  })
  return data
}
