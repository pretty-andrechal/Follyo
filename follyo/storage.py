"""JSON-based storage for portfolio data."""

import json
from pathlib import Path
from typing import List, Optional

from .models import Holding, Loan, Sale


DEFAULT_DATA_PATH = Path(__file__).parent.parent / "data" / "portfolio.json"


class PortfolioStorage:
    """Handles persistence of portfolio data to JSON."""

    def __init__(self, data_path: Optional[Path] = None):
        self.data_path = data_path or DEFAULT_DATA_PATH
        self._ensure_data_file()

    def _ensure_data_file(self):
        """Create data file if it doesn't exist."""
        self.data_path.parent.mkdir(parents=True, exist_ok=True)
        if not self.data_path.exists():
            self._save_data({"holdings": [], "loans": [], "sales": []})

    def _load_data(self) -> dict:
        """Load data from JSON file."""
        with open(self.data_path, "r") as f:
            return json.load(f)

    def _save_data(self, data: dict):
        """Save data to JSON file."""
        with open(self.data_path, "w") as f:
            json.dump(data, f, indent=2)

    # Holdings operations
    def get_holdings(self) -> List[Holding]:
        """Get all holdings."""
        data = self._load_data()
        return [Holding.from_dict(h) for h in data.get("holdings", [])]

    def add_holding(self, holding: Holding) -> Holding:
        """Add a new holding."""
        data = self._load_data()
        data["holdings"].append(holding.to_dict())
        self._save_data(data)
        return holding

    def remove_holding(self, holding_id: str) -> bool:
        """Remove a holding by ID."""
        data = self._load_data()
        original_len = len(data["holdings"])
        data["holdings"] = [h for h in data["holdings"] if h["id"] != holding_id]
        if len(data["holdings"]) < original_len:
            self._save_data(data)
            return True
        return False

    # Loans operations
    def get_loans(self) -> List[Loan]:
        """Get all loans."""
        data = self._load_data()
        return [Loan.from_dict(l) for l in data.get("loans", [])]

    def add_loan(self, loan: Loan) -> Loan:
        """Add a new loan."""
        data = self._load_data()
        data["loans"].append(loan.to_dict())
        self._save_data(data)
        return loan

    def remove_loan(self, loan_id: str) -> bool:
        """Remove a loan by ID."""
        data = self._load_data()
        original_len = len(data["loans"])
        data["loans"] = [l for l in data["loans"] if l["id"] != loan_id]
        if len(data["loans"]) < original_len:
            self._save_data(data)
            return True
        return False

    # Sales operations
    def get_sales(self) -> List[Sale]:
        """Get all sales."""
        data = self._load_data()
        return [Sale.from_dict(s) for s in data.get("sales", [])]

    def add_sale(self, sale: Sale) -> Sale:
        """Add a new sale."""
        data = self._load_data()
        if "sales" not in data:
            data["sales"] = []
        data["sales"].append(sale.to_dict())
        self._save_data(data)
        return sale

    def remove_sale(self, sale_id: str) -> bool:
        """Remove a sale by ID."""
        data = self._load_data()
        original_len = len(data.get("sales", []))
        data["sales"] = [s for s in data.get("sales", []) if s["id"] != sale_id]
        if len(data["sales"]) < original_len:
            self._save_data(data)
            return True
        return False
