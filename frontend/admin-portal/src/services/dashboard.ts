/**
 * Dashboard API 服务
 *
 * 注意: Admin Portal的Dashboard数据需要从多个服务聚合
 * 后端/api/v1/dashboard是为Merchant Portal设计的
 * Admin Portal需要聚合以下API:
 * - /api/v1/merchant (商户统计)
 * - /api/v1/admin (管理员统计)
 * - /api/v1/orders/stats (订单统计)
 * - /api/v1/payments (支付统计)
 * - /api/v1/transactions/stats (交易统计)
 */
import request from './request'
import type { DashboardData } from '@payment/shared/types'

/**
 * 获取Dashboard数据 (聚合多个API)
 * @param timePeriod 时间周期: today | 7days | 30days
 */
export const getDashboardData = async (
  timePeriod: 'today' | '7days' | '30days' = 'today',
): Promise<DashboardData> => {
  try {
    // 聚合多个API的数据
    const [merchantsResponse, adminsResponse, ordersStatsResponse, paymentsResponse] =
      await Promise.all([
        request.get('/api/v1/merchant', { params: { page: 1, page_size: 1 } }),
        request.get('/api/v1/admin', { params: { page: 1, page_size: 1 } }),
        request.get('/api/v1/orders/stats', { params: { period: timePeriod } }),
        request.get('/api/v1/payments', { params: { page: 1, page_size: 10 } }),
      ])

    // 组装Dashboard数据
    const dashboardData: DashboardData = {
      total_merchants: merchantsResponse.pagination?.total || 0,
      total_admins: adminsResponse.pagination?.total || 0,
      total_orders: ordersStatsResponse.total_count || 0,
      total_amount: ordersStatsResponse.total_amount || 0,
      today_orders: ordersStatsResponse.today_count || 0,
      today_amount: ordersStatsResponse.today_amount || 0,
      recent_payments: paymentsResponse.list || [],
    }

    return dashboardData
  } catch (error) {
    console.error('Failed to fetch dashboard data:', error)
    throw error
  }
}

/**
 * 获取Dashboard统计数据
 */
export const getDashboardStats = async () => {
  try {
    const [ordersStats, transactionsStats] = await Promise.all([
      request.get('/api/v1/orders/stats'),
      request.get('/api/v1/transactions/stats'),
    ])

    return {
      orders: ordersStats,
      transactions: transactionsStats,
    }
  } catch (error) {
    console.error('Failed to fetch dashboard stats:', error)
    throw error
  }
}

/**
 * 获取交易趋势数据 - 注意: 后端需要实现此接口
 */
export const getTrendData = async (period: 'today' | '7days' | '30days' = 'today') => {
  const { data } = await request.get('/api/v1/analytics/trend', {
    params: { period },
  })
  return data
}

/**
 * 获取渠道分布数据 - 注意: 后端需要实现此接口
 */
export const getChannelDistribution = async () => {
  const { data } = await request.get('/api/v1/analytics/channel-distribution')
  return data
}

/**
 * 获取商户排行 - 注意: 后端需要实现此接口
 */
export const getMerchantRanks = async (limit = 5) => {
  const { data } = await request.get('/api/v1/analytics/merchant-ranks', {
    params: { limit },
  })
  return data
}

/**
 * 获取近期活动 - 从审计日志获取
 */
export const getRecentActivities = async (limit = 5) => {
  const { data } = await request.get('/api/v1/audit-logs', {
    params: { page: 1, page_size: limit },
  })
  return data
}
