// ==================== 通用类型 ====================

export interface ApiResponse<T = any> {
  code: number
  message: string
  data?: T
  error?: {
    code: string
    message: string
    details?: any
  }
}

export interface PaginationParams {
  page: number
  page_size: number
}

export interface PaginationResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// ==================== 管理员相关 ====================

export interface Permission {
  id: string
  code: string
  name: string
  resource: string
  action: string
  description?: string
  created_at: string
  updated_at: string
}

export interface Role {
  id: string
  name: string
  display_name: string
  description: string
  permissions: Permission[]
  is_system?: boolean
  created_at: string
  updated_at: string
}

export interface Admin {
  id: string
  username: string
  email: string
  full_name: string
  avatar: string
  status: 'active' | 'inactive' | 'locked'
  is_super: boolean
  roles: Role[]
  last_login_at?: string
  last_login_ip?: string
  created_at: string
  updated_at: string
}

// ==================== 商户相关 ====================

export interface Merchant {
  id: string
  name: string
  email: string
  phone?: string
  country: string
  business_type: string
  website?: string
  status: 'pending' | 'active' | 'frozen' | 'rejected'
  kyc_status: 'pending' | 'approved' | 'rejected'
  balance: number
  frozen_balance: number
  created_at: string
  updated_at: string
}

export interface ApiKey {
  id: string
  merchant_id: string
  key_id: string
  key_secret?: string // 只在创建时返回
  name: string
  status: 'active' | 'inactive'
  expires_at?: string
  last_used_at?: string
  created_at: string
  updated_at: string
}

export interface Webhook {
  id: string
  merchant_id: string
  url: string
  events: string[]
  status: 'active' | 'inactive'
  secret: string
  retry_count: number
  created_at: string
  updated_at: string
}

// ==================== 支付相关 ====================

export interface Payment {
  id: string
  payment_no: string
  merchant_id: string
  merchant_order_no: string
  amount: number
  currency: string
  channel: string
  channel_trade_no?: string
  status: 'pending' | 'processing' | 'success' | 'failed' | 'cancelled'
  payment_method?: string
  payment_url?: string
  client_ip?: string
  subject?: string
  description?: string
  return_url?: string
  notify_url?: string
  extra?: Record<string, any>
  paid_at?: string
  expired_at?: string
  created_at: string
  updated_at: string
}

export interface Refund {
  id: string
  refund_no: string
  payment_id: string
  payment_no: string
  merchant_id: string
  amount: number
  currency: string
  reason?: string
  status: 'pending' | 'processing' | 'success' | 'failed'
  channel_refund_no?: string
  refunded_at?: string
  created_at: string
  updated_at: string
}

// ==================== 订单相关 ====================

export interface Order {
  id: string
  order_no: string
  merchant_id: string
  merchant_order_no: string
  amount: number
  currency: string
  status: 'pending' | 'paid' | 'cancelled' | 'refunded' | 'partial_refunded'
  paid_amount: number
  refund_amount: number
  subject: string
  description?: string
  extra?: Record<string, any>
  paid_at?: string
  expired_at?: string
  created_at: string
  updated_at: string
}

// ==================== 风控相关 ====================

export interface RiskRule {
  id: string
  name: string
  type: 'amount_limit' | 'frequency_limit' | 'ip_blacklist' | 'card_blacklist' | 'merchant_blacklist'
  conditions: Record<string, any>
  action: 'allow' | 'reject' | 'review'
  priority: number
  status: 'active' | 'inactive'
  created_at: string
  updated_at: string
}

export interface RiskEvent {
  id: string
  merchant_id: string
  payment_id?: string
  event_type: string
  risk_level: 'low' | 'medium' | 'high' | 'critical'
  score: number
  description: string
  details?: Record<string, any>
  status: 'pending' | 'reviewed' | 'ignored'
  reviewed_by?: string
  reviewed_at?: string
  created_at: string
}

// ==================== 账务相关 ====================

export interface Account {
  id: string
  merchant_id: string
  balance: number
  frozen_balance: number
  total_in: number
  total_out: number
  currency: string
  created_at: string
  updated_at: string
}

export interface Settlement {
  id: string
  settlement_no: string
  merchant_id: string
  amount: number
  fee: number
  actual_amount: number
  currency: string
  status: 'pending' | 'processing' | 'success' | 'failed'
  transaction_count: number
  start_time: string
  end_time: string
  settled_at?: string
  created_at: string
  updated_at: string
}

// ==================== 系统配置 ====================

export interface SystemConfig {
  id: string
  category: string
  key: string
  value: string
  value_type: 'string' | 'number' | 'boolean' | 'json'
  description?: string
  is_public: boolean
  is_encrypted: boolean
  created_at: string
  updated_at: string
}

export interface AuditLog {
  id: string
  admin_id: string
  admin_username: string
  action: string
  resource: string
  resource_id?: string
  details?: Record<string, any>
  ip_address: string
  user_agent?: string
  status: 'success' | 'failed'
  error_message?: string
  created_at: string
}

// ==================== 统计数据 ====================

export interface DashboardStats {
  total_merchants: number
  active_merchants: number
  total_transactions: number
  total_amount: number
  success_rate: number
  pending_orders: number
  today_growth: number
  total_admins: number
}

export interface ChartData {
  date: string
  value: number
}

export interface ChannelDistribution {
  type: string
  value: number
}


