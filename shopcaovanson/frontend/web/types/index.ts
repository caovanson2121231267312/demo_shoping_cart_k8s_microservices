export interface User {
  id: string
  email: string
  full_name: string
  role: UserRole
  is_active?: boolean
  email_verified?: boolean
  has_avatar?: boolean
  avatar_url?: string
  created_at: string
}

export interface RegisterResponse {
  message: string
  email: string
  requires_verification: boolean
}

export type UserRole =
  | 'super_admin'
  | 'admin'
  | 'manager'
  | 'staff'
  | 'support'
  | 'customer'

export interface TokenPair {
  access_token: string
  refresh_token: string
}

export interface RefreshResponse {
  access_token: string
}

export interface Category {
  id: string
  name: string
  slug: string
  parent_id?: string | null
  image_url?: string | null
  children?: Category[]
}

export interface Product {
  id: string
  category_id: string
  name: string
  slug: string
  description?: string | null
  price: number
  sale_price?: number | null
  stock: number
  images: string[]
  is_active: boolean
  created_at: string
  updated_at: string
  category?: Category | null
  details?: ProductDetail | null
}

export interface ProductDetail {
  product_id: string
  specifications: Record<string, string>
  rich_description: string
  updated_at: string
}

export interface ProductListResult {
  items: Product[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export interface ProductReview {
  id: string
  product_id: string
  user_id: string
  rating: number
  comment?: string | null
  created_at: string
}

export interface CreateReviewInput {
  rating: number
  comment?: string
}

export interface ReviewListResult {
  items: ProductReview[]
  total: number
}

export interface CartItem {
  product_id: string
  quantity: number
  unit_price: number
  product_name: string
  product_image?: string
}

export interface Cart {
  user_id: string
  items: CartItem[]
  updated_at: string
}

export interface OrderItem {
  id: string
  order_id: string
  product_id: string
  product_name_snapshot: string
  product_image_snapshot?: string | null
  unit_price: number
  quantity: number
}

export interface Order {
  id: string
  order_number?: string
  user_id: string
  status: 'pending' | 'confirmed' | 'shipping' | 'delivered' | 'cancelled'
  subtotal_amount?: number
  discount_amount?: number
  coupon_code?: string | null
  total_amount: number
  shipping_name: string
  shipping_phone: string
  shipping_address: string
  created_at: string
  updated_at: string
  items?: OrderItem[]
}

export interface OrderListResult {
  items: Order[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export interface ChatRoom {
  id: string
  participants: string[]
  room_type: string
  created_at: string
}

export interface ChatProductSuggestion {
  id: string
  name: string
  slug: string
  price: number
  sale_price?: number | null
  image_url?: string | null
  stock?: number
}

export interface ChatMessage {
  id: string
  room_id: string
  sender_id: string
  content: string
  type: string
  products?: ChatProductSuggestion[]
  reactions?: Record<string, string[]>
  read_by?: string[]
  created_at: string
}

export interface MessageListResult {
  items: ChatMessage[]
  next_cursor?: string | null
}

export interface WSClientMessage {
  type: 'join' | 'message' | 'typing' | 'reaction'
  room_id: string
  content?: string
  msg_type?: 'text' | 'image'
  message_id?: string
  emoji?: string
}

export interface WSServerMessage {
  type: 'message' | 'user_joined' | 'joined' | 'typing' | 'reaction'
  room_id?: string
  message_id?: string
  sender_id?: string
  user_id?: string
  content?: string
  msg_type?: string
  emoji?: string
  reactions?: Record<string, string[]>
  created_at?: string
}

export interface ApiError {
  error: string
  message: string
}

export interface ProductFilters {
  page?: number
  limit?: number
  category?: string
  search?: string
  min_price?: number
  max_price?: number
  sort?: 'price_asc' | 'price_desc' | 'newest'
  include_inactive?: boolean
  created_from?: string
  created_to?: string
}

export interface CheckoutInput {
  shipping_name: string
  shipping_phone: string
  shipping_address: string
  coupon_code?: string
}

export interface Coupon {
  id: string
  code: string
  type: 'percent' | 'fixed'
  value: number
  min_order: number
  max_discount?: number | null
  usage_limit?: number | null
  used_count: number
  is_active: boolean
  expires_at?: string | null
  created_at: string
  updated_at: string
}

export interface CouponListResult {
  items: Coupon[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export interface ValidateCouponResult {
  valid: boolean
  code?: string
  discount_amount: number
  final_amount: number
  message?: string
}

export interface TrackOrderInput {
  order_number: string
  shipping_phone: string
}

export interface UserListResult {
  items: User[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export interface AdminUserStats {
  total_users: number
  users_by_role: Record<string, number>
  active_users: number
  inactive_users: number
}

export interface OrderStats {
  total_orders: number
  revenue: number
  by_status: Record<string, number>
}

export interface RoleInfo {
  id: string
  label: string
  level: number
  description: string
}

export type AnalyticsPeriod = 'day' | 'week' | 'month' | 'quarter'

export interface AnalyticsOverview {
  period: AnalyticsPeriod
  from: string
  to: string
  granularity: string
  orders: {
    total: number
    revenue: number
    pending: number
    delivered: number
    by_status: Record<string, number>
  }
  users: {
    new_customers: number
    total_customers: number
    active_users: number
  }
  products: {
    total_products: number
    new_reviews: number
    avg_rating: number
    total_reviews: number
  }
}

export interface AnalyticsSeriesPoint {
  label: string
  revenue?: number
  orders?: number
  count?: number
  avg_rating?: number
}

export interface TopProductRow {
  name: string
  quantity_sold: number
  revenue: number
  order_count: number
}

export interface OnlineUser {
  user_id: string
  email: string
  page: string
  last_seen: number
  role?: string
}

export interface SupportChatRoom extends ChatRoom {
  customer_id: string
}

export interface Article {
  id: string
  title: string
  slug: string
  excerpt?: string | null
  content: string
  cover_image?: string | null
  category: string
  tags: string[]
  author_name: string
  is_published: boolean
  is_featured: boolean
  view_count: number
  created_at: string
  updated_at: string
}

export interface ArticleListResult {
  items: Article[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export const ARTICLE_CATEGORIES: Record<string, string> = {
  'tin-tuc': 'Tin tức',
  'huong-dan': 'Hướng dẫn mua sắm',
  'khuyen-mai': 'Khuyến mãi',
  'danh-gia': 'Đánh giá sản phẩm',
  'xu-huong': 'Xu hướng',
  'cong-nghe': 'Công nghệ',
  'meo-vat': 'Mẹo vặt',
}

export type CaroRoomStatus = 'waiting' | 'rps' | 'playing' | 'finished'
export type CaroSymbol = 'X' | 'O' | null

export interface CaroRoom {
  id: string
  code: string
  board_size: number
  status: CaroRoomStatus
  visibility: 'public' | 'private'
  host_name?: string
  player_x_name?: string | null
  player_o_name?: string | null
  player_x_id?: string | null
  player_o_id?: string | null
  current_turn?: CaroSymbol
  winner_id?: string | null
  you_are?: 'X' | 'O' | null
  is_host?: boolean
}

export interface CaroChatMessage {
  id?: string
  user_id: string
  user_name: string
  content: string
  created_at?: string
}

export interface CaroPublicRoom {
  id: string
  code: string
  board_size: number
  status: CaroRoomStatus
  host_name: string
  player_x_name?: string | null
  player_o_name?: string | null
  created_at: string
  updated_at?: string
  visibility?: 'public' | 'private'
}
