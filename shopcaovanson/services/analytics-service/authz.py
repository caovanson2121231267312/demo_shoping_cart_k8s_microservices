from fastapi import Header, HTTPException

ROLE_LEVEL = {
    "customer": 0,
    "support": 20,
    "staff": 40,
    "manager": 60,
    "admin": 80,
    "super_admin": 100,
}


def require_min_role(min_role: str):
    min_level = ROLE_LEVEL.get(min_role, 100)

    def checker(x_user_role: str | None = Header(default=None, alias="X-User-Role")):
        if not x_user_role:
            raise HTTPException(status_code=401, detail="authentication required")
        level = ROLE_LEVEL.get(x_user_role, 0)
        if level < min_level:
            raise HTTPException(status_code=403, detail="insufficient permissions")
        return x_user_role

    return checker


def get_user_id(x_user_id: str | None = Header(default=None, alias="X-User-Id")) -> str | None:
    return x_user_id
