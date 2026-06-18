export type UserRole =
  | 'super_admin'
  | 'admin'
  | 'manager'
  | 'staff'
  | 'support'
  | 'customer'

const roleLevel: Record<UserRole, number> = {
  super_admin: 100,
  admin: 80,
  manager: 60,
  staff: 40,
  support: 20,
  customer: 0,
}

export function roleLevelOf(role: string): number {
  return roleLevel[role as UserRole] ?? 0
}

export function hasMinRole(role: string, min: UserRole): boolean {
  return roleLevelOf(role) >= roleLevel[min]
}

export const STAFF_ROLES: UserRole[] = ['super_admin', 'admin', 'manager', 'staff', 'support']
