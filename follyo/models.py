"""Data models for portfolio tracking."""

from dataclasses import dataclass, asdict
from datetime import datetime
from typing import Optional
import uuid


@dataclass
class Holding:
    """Represents a crypto holding/purchase."""
    id: str
    coin: str
    amount: float
    purchase_price_usd: float
    date: str
    platform: Optional[str] = None
    notes: Optional[str] = None

    @classmethod
    def create(
        cls,
        coin: str,
        amount: float,
        purchase_price_usd: float,
        platform: Optional[str] = None,
        notes: Optional[str] = None,
        date: Optional[str] = None
    ) -> "Holding":
        """Create a new holding with auto-generated ID and date."""
        return cls(
            id=str(uuid.uuid4())[:8],
            coin=coin.upper(),
            amount=amount,
            purchase_price_usd=purchase_price_usd,
            date=date or datetime.now().strftime("%Y-%m-%d"),
            platform=platform,
            notes=notes
        )

    def to_dict(self) -> dict:
        return asdict(self)

    @classmethod
    def from_dict(cls, data: dict) -> "Holding":
        return cls(**data)

    @property
    def total_value_usd(self) -> float:
        """Total value at purchase price."""
        return self.amount * self.purchase_price_usd


@dataclass
class Loan:
    """Represents a crypto loan on a platform."""
    id: str
    coin: str
    amount: float
    platform: str
    date: str
    interest_rate: Optional[float] = None  # Annual percentage
    notes: Optional[str] = None

    @classmethod
    def create(
        cls,
        coin: str,
        amount: float,
        platform: str,
        interest_rate: Optional[float] = None,
        notes: Optional[str] = None,
        date: Optional[str] = None
    ) -> "Loan":
        """Create a new loan with auto-generated ID and date."""
        return cls(
            id=str(uuid.uuid4())[:8],
            coin=coin.upper(),
            amount=amount,
            platform=platform,
            date=date or datetime.now().strftime("%Y-%m-%d"),
            interest_rate=interest_rate,
            notes=notes
        )

    def to_dict(self) -> dict:
        return asdict(self)

    @classmethod
    def from_dict(cls, data: dict) -> "Loan":
        return cls(**data)
