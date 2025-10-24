export interface CashierSession {
  id: string
  session_token: string
  merchant_id: string
  order_no: string
  amount: number
  currency: string
  description: string
  customer_email: string
  customer_name: string
  customer_ip: string
  allowed_channels: string[]
  allowed_methods: string[]
  status: 'pending' | 'completed' | 'expired' | 'cancelled'
  payment_no?: string
  expires_at: string
  completed_at?: string
  metadata?: Record<string, any>
  created_at: string
  updated_at: string
}

export interface CashierConfig {
  id: string
  merchant_id: string
  theme_color: string
  logo_url: string
  background_image_url: string
  custom_css: string
  enabled_channels: string[]
  default_channel: string
  enabled_languages: string[]
  default_language: string
  auto_submit: boolean
  show_amount_breakdown: boolean
  allow_channel_switch: boolean
  session_timeout_minutes: number
  require_cvv: boolean
  enable_3d_secure: boolean
  allowed_countries: string[]
  success_redirect_url: string
  cancel_redirect_url: string
  created_at: string
  updated_at: string
}

export interface PaymentMethod {
  id: string
  name: string
  icon: string
  enabled: boolean
}

export interface PaymentFormData {
  cardNumber: string
  cardholderName: string
  expiryDate: string
  cvv: string
  email: string
  saveCard: boolean
}
