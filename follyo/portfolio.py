"""Portfolio management logic."""

from typing import Dict, List, Optional
from collections import defaultdict

from .models import Holding, Loan
from .storage import PortfolioStorage


class Portfolio:
    """Manages crypto holdings and loans."""

    def __init__(self, storage: Optional[PortfolioStorage] = None):
        self.storage = storage or PortfolioStorage()

    # Holdings
    def add_holding(
        self,
        coin: str,
        amount: float,
        purchase_price_usd: float,
        platform: Optional[str] = None,
        notes: Optional[str] = None,
        date: Optional[str] = None
    ) -> Holding:
        """Add a new coin holding."""
        holding = Holding.create(
            coin=coin,
            amount=amount,
            purchase_price_usd=purchase_price_usd,
            platform=platform,
            notes=notes,
            date=date
        )
        return self.storage.add_holding(holding)

    def remove_holding(self, holding_id: str) -> bool:
        """Remove a holding by ID."""
        return self.storage.remove_holding(holding_id)

    def list_holdings(self) -> List[Holding]:
        """List all holdings."""
        return self.storage.get_holdings()

    # Loans
    def add_loan(
        self,
        coin: str,
        amount: float,
        platform: str,
        interest_rate: Optional[float] = None,
        notes: Optional[str] = None,
        date: Optional[str] = None
    ) -> Loan:
        """Add a new loan."""
        loan = Loan.create(
            coin=coin,
            amount=amount,
            platform=platform,
            interest_rate=interest_rate,
            notes=notes,
            date=date
        )
        return self.storage.add_loan(loan)

    def remove_loan(self, loan_id: str) -> bool:
        """Remove a loan by ID."""
        return self.storage.remove_loan(loan_id)

    def list_loans(self) -> List[Loan]:
        """List all loans."""
        return self.storage.get_loans()

    # Summary
    def get_holdings_by_coin(self) -> Dict[str, float]:
        """Get total holdings aggregated by coin."""
        holdings = self.list_holdings()
        by_coin = defaultdict(float)
        for h in holdings:
            by_coin[h.coin] += h.amount
        return dict(by_coin)

    def get_loans_by_coin(self) -> Dict[str, float]:
        """Get total loans aggregated by coin."""
        loans = self.list_loans()
        by_coin = defaultdict(float)
        for l in loans:
            by_coin[l.coin] += l.amount
        return dict(by_coin)

    def get_net_holdings_by_coin(self) -> Dict[str, float]:
        """Get net holdings (holdings - loans) by coin."""
        holdings = self.get_holdings_by_coin()
        loans = self.get_loans_by_coin()

        all_coins = set(holdings.keys()) | set(loans.keys())
        net = {}
        for coin in all_coins:
            net[coin] = holdings.get(coin, 0) - loans.get(coin, 0)
        return net

    def get_total_invested_usd(self) -> float:
        """Get total USD invested in holdings."""
        return sum(h.total_value_usd for h in self.list_holdings())

    def get_summary(self) -> dict:
        """Get portfolio summary."""
        holdings = self.list_holdings()
        loans = self.list_loans()

        return {
            "total_holdings_count": len(holdings),
            "total_loans_count": len(loans),
            "total_invested_usd": self.get_total_invested_usd(),
            "holdings_by_coin": self.get_holdings_by_coin(),
            "loans_by_coin": self.get_loans_by_coin(),
            "net_by_coin": self.get_net_holdings_by_coin(),
        }
