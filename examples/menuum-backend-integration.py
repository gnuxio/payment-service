"""
Example integration code for menuum-backend (Python/FastAPI)
This shows how to integrate with payment-service
"""

import requests
from fastapi import APIRouter, Request, HTTPException
from typing import Optional

# Configuration
PAYMENT_SERVICE_URL = "http://localhost:8081"
PAYMENT_SERVICE_API_KEY = "your_api_key_here"  # Same as in payment-service .env
TENANT_ID = "menuum"

router = APIRouter()


# ============================================================================
# 1. CREATE CHECKOUT SESSION
# ============================================================================

@router.post("/subscriptions/create-checkout")
async def create_checkout_session(user_id: str, plan: str):
    """
    Create a Stripe checkout session for a user

    Args:
        user_id: AWS Cognito user ID
        plan: "premium_monthly" or "premium_yearly"
    """
    headers = {
        "X-API-Key": PAYMENT_SERVICE_API_KEY,
        "X-Tenant-ID": TENANT_ID,
        "Content-Type": "application/json"
    }

    payload = {
        "user_id": user_id,
        "plan": plan,
        "success_url": "https://menuum.com/subscription/success",
        "cancel_url": "https://menuum.com/subscription/cancel"
    }

    try:
        response = requests.post(
            f"{PAYMENT_SERVICE_URL}/payments/checkout",
            json=payload,
            headers=headers
        )
        response.raise_for_status()

        data = response.json()
        # Return the session URL to redirect user
        return {
            "checkout_url": data["session_url"],
            "session_id": data["session_id"]
        }
    except requests.exceptions.RequestException as e:
        raise HTTPException(status_code=500, detail=f"Error creating checkout: {str(e)}")


# ============================================================================
# 2. GET SUBSCRIPTION STATUS
# ============================================================================

@router.get("/subscriptions/status/{user_id}")
async def get_subscription_status(user_id: str):
    """
    Get the subscription status for a user

    Args:
        user_id: AWS Cognito user ID
    """
    headers = {
        "X-API-Key": PAYMENT_SERVICE_API_KEY,
        "X-Tenant-ID": TENANT_ID
    }

    try:
        response = requests.get(
            f"{PAYMENT_SERVICE_URL}/payments/subscription/{user_id}",
            headers=headers
        )

        if response.status_code == 404:
            return {"has_subscription": False}

        response.raise_for_status()
        subscription = response.json()

        return {
            "has_subscription": True,
            "status": subscription["status"],
            "plan": subscription["plan"],
            "is_active": subscription["status"] == "active",
            "current_period_end": subscription.get("current_period_end"),
            "cancel_at_period_end": subscription.get("cancel_at_period_end", False)
        }
    except requests.exceptions.RequestException as e:
        raise HTTPException(status_code=500, detail=f"Error fetching subscription: {str(e)}")


# ============================================================================
# 3. CANCEL SUBSCRIPTION
# ============================================================================

@router.post("/subscriptions/cancel/{user_id}")
async def cancel_subscription(user_id: str):
    """
    Cancel a user's subscription

    Args:
        user_id: AWS Cognito user ID
    """
    headers = {
        "X-API-Key": PAYMENT_SERVICE_API_KEY,
        "X-Tenant-ID": TENANT_ID
    }

    try:
        response = requests.post(
            f"{PAYMENT_SERVICE_URL}/payments/cancel/{user_id}",
            headers=headers
        )
        response.raise_for_status()

        return response.json()
    except requests.exceptions.RequestException as e:
        raise HTTPException(status_code=500, detail=f"Error canceling subscription: {str(e)}")


# ============================================================================
# 4. WEBHOOK ENDPOINT (receives notifications from payment-service)
# ============================================================================

@router.post("/webhooks/subscription")
async def subscription_webhook(request: Request):
    """
    Receives webhook notifications from payment-service when subscription status changes
    This endpoint MUST exist for payment-service to notify menuum-backend
    """
    payload = await request.json()

    user_id = payload.get("user_id")
    status = payload.get("status")
    plan = payload.get("plan")
    subscription_id = payload.get("subscription_id")

    print(f"Received webhook for user {user_id}: status={status}, plan={plan}")

    # Update user's premium status in your database
    if status == "active":
        # User's subscription is active - grant premium access
        await update_user_premium_status(user_id, is_premium=True)
    elif status in ["canceled", "incomplete_expired", "unpaid"]:
        # User's subscription is not active - revoke premium access
        await update_user_premium_status(user_id, is_premium=False)

    return {"status": "success"}


# ============================================================================
# HELPER FUNCTIONS
# ============================================================================

async def update_user_premium_status(user_id: str, is_premium: bool):
    """
    Update the is_premium field for a user in your database

    Args:
        user_id: AWS Cognito user ID
        is_premium: Whether user should have premium access
    """
    # Example with SQLAlchemy
    from sqlalchemy import update
    from your_db_models import User
    from your_db import get_db

    db = get_db()

    stmt = (
        update(User)
        .where(User.cognito_user_id == user_id)
        .values(is_premium=is_premium)
    )

    await db.execute(stmt)
    await db.commit()

    print(f"Updated user {user_id} premium status to {is_premium}")


async def get_user_premium_status(user_id: str) -> bool:
    """
    Get the premium status for a user from your database

    Args:
        user_id: AWS Cognito user ID

    Returns:
        bool: Whether user has premium access
    """
    # Example with SQLAlchemy
    from sqlalchemy import select
    from your_db_models import User
    from your_db import get_db

    db = get_db()

    result = await db.execute(
        select(User.is_premium)
        .where(User.cognito_user_id == user_id)
    )

    row = result.first()
    return row[0] if row else False


# ============================================================================
# MIDDLEWARE (Optional) - Check premium status on protected routes
# ============================================================================

from fastapi import Depends, HTTPException
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials

security = HTTPBearer()

async def require_premium(credentials: HTTPAuthorizationCredentials = Depends(security)):
    """
    Middleware to require premium subscription for certain routes
    """
    # Extract user_id from token (example with JWT)
    # You would decode the JWT token to get the user_id
    user_id = "user_id_from_token"  # Replace with actual token decoding

    is_premium = await get_user_premium_status(user_id)

    if not is_premium:
        raise HTTPException(
            status_code=403,
            detail="Premium subscription required"
        )

    return user_id


# Usage example:
@router.get("/premium-feature", dependencies=[Depends(require_premium)])
async def premium_feature():
    """
    This route is only accessible to premium users
    """
    return {"message": "This is a premium feature!"}
