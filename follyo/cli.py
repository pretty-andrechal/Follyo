"""Command-line interface for Follyo portfolio tracker."""

import click
from tabulate import tabulate

from .portfolio import Portfolio


@click.group()
@click.pass_context
def cli(ctx):
    """Follyo - Personal Crypto Portfolio Tracker

    Track your crypto holdings, sales, and loans across platforms.
    """
    ctx.ensure_object(dict)
    ctx.obj["portfolio"] = Portfolio()


# ============ Holdings Commands ============

@cli.group()
def holding():
    """Manage crypto holdings."""
    pass


@holding.command("add")
@click.argument("coin")
@click.argument("amount", type=float)
@click.argument("price", type=float)
@click.option("-p", "--platform", help="Platform where held (e.g., Binance, Ledger)")
@click.option("-n", "--notes", help="Optional notes")
@click.option("-d", "--date", help="Purchase date (YYYY-MM-DD)")
@click.pass_context
def add_holding(ctx, coin, amount, price, platform, notes, date):
    """Add a coin holding.

    COIN: The cryptocurrency symbol (e.g., BTC, ETH)
    AMOUNT: Amount of coins
    PRICE: Purchase price per coin in USD
    """
    portfolio = ctx.obj["portfolio"]
    holding = portfolio.add_holding(
        coin=coin,
        amount=amount,
        purchase_price_usd=price,
        platform=platform,
        notes=notes,
        date=date
    )
    click.echo(f"Added holding: {holding.amount} {holding.coin} @ ${holding.purchase_price_usd} (ID: {holding.id})")


@holding.command("list")
@click.pass_context
def list_holdings(ctx):
    """List all holdings."""
    portfolio = ctx.obj["portfolio"]
    holdings = portfolio.list_holdings()

    if not holdings:
        click.echo("No holdings found.")
        return

    table_data = []
    for h in holdings:
        table_data.append([
            h.id,
            h.coin,
            f"{h.amount:,.8f}".rstrip('0').rstrip('.'),
            f"${h.purchase_price_usd:,.2f}",
            f"${h.total_value_usd:,.2f}",
            h.platform or "-",
            h.date
        ])

    headers = ["ID", "Coin", "Amount", "Price/Unit", "Total USD", "Platform", "Date"]
    click.echo(tabulate(table_data, headers=headers, tablefmt="simple"))


@holding.command("remove")
@click.argument("holding_id")
@click.pass_context
def remove_holding(ctx, holding_id):
    """Remove a holding by ID."""
    portfolio = ctx.obj["portfolio"]
    if portfolio.remove_holding(holding_id):
        click.echo(f"Removed holding {holding_id}")
    else:
        click.echo(f"Holding {holding_id} not found")


# ============ Loan Commands ============

@cli.group()
def loan():
    """Manage crypto loans."""
    pass


@loan.command("add")
@click.argument("coin")
@click.argument("amount", type=float)
@click.argument("platform")
@click.option("-r", "--rate", type=float, help="Annual interest rate (%)")
@click.option("-n", "--notes", help="Optional notes")
@click.option("-d", "--date", help="Loan date (YYYY-MM-DD)")
@click.pass_context
def add_loan(ctx, coin, amount, platform, rate, notes, date):
    """Add a loan.

    COIN: The cryptocurrency symbol (e.g., BTC, USDT)
    AMOUNT: Amount borrowed
    PLATFORM: Platform where loan is held (e.g., Nexo, Celsius)
    """
    portfolio = ctx.obj["portfolio"]
    loan = portfolio.add_loan(
        coin=coin,
        amount=amount,
        platform=platform,
        interest_rate=rate,
        notes=notes,
        date=date
    )
    click.echo(f"Added loan: {loan.amount} {loan.coin} on {loan.platform} (ID: {loan.id})")


@loan.command("list")
@click.pass_context
def list_loans(ctx):
    """List all loans."""
    portfolio = ctx.obj["portfolio"]
    loans = portfolio.list_loans()

    if not loans:
        click.echo("No loans found.")
        return

    table_data = []
    for l in loans:
        rate_str = f"{l.interest_rate}%" if l.interest_rate else "-"
        table_data.append([
            l.id,
            l.coin,
            f"{l.amount:,.8f}".rstrip('0').rstrip('.'),
            l.platform,
            rate_str,
            l.date
        ])

    headers = ["ID", "Coin", "Amount", "Platform", "Rate", "Date"]
    click.echo(tabulate(table_data, headers=headers, tablefmt="simple"))


@loan.command("remove")
@click.argument("loan_id")
@click.pass_context
def remove_loan(ctx, loan_id):
    """Remove a loan by ID."""
    portfolio = ctx.obj["portfolio"]
    if portfolio.remove_loan(loan_id):
        click.echo(f"Removed loan {loan_id}")
    else:
        click.echo(f"Loan {loan_id} not found")


# ============ Sale Commands ============

@cli.group()
def sale():
    """Manage crypto sales."""
    pass


@sale.command("add")
@click.argument("coin")
@click.argument("amount", type=float)
@click.argument("price", type=float)
@click.option("-p", "--platform", help="Platform where sold (e.g., Binance, Kraken)")
@click.option("-n", "--notes", help="Optional notes")
@click.option("-d", "--date", help="Sale date (YYYY-MM-DD)")
@click.pass_context
def add_sale(ctx, coin, amount, price, platform, notes, date):
    """Add a coin sale.

    COIN: The cryptocurrency symbol (e.g., BTC, ETH)
    AMOUNT: Amount of coins sold
    PRICE: Sell price per coin in USD
    """
    portfolio = ctx.obj["portfolio"]
    sale = portfolio.add_sale(
        coin=coin,
        amount=amount,
        sell_price_usd=price,
        platform=platform,
        notes=notes,
        date=date
    )
    click.echo(f"Added sale: {sale.amount} {sale.coin} @ ${sale.sell_price_usd} (ID: {sale.id})")


@sale.command("list")
@click.pass_context
def list_sales(ctx):
    """List all sales."""
    portfolio = ctx.obj["portfolio"]
    sales = portfolio.list_sales()

    if not sales:
        click.echo("No sales found.")
        return

    table_data = []
    for s in sales:
        table_data.append([
            s.id,
            s.coin,
            f"{s.amount:,.8f}".rstrip('0').rstrip('.'),
            f"${s.sell_price_usd:,.2f}",
            f"${s.total_value_usd:,.2f}",
            s.platform or "-",
            s.date
        ])

    headers = ["ID", "Coin", "Amount", "Price/Unit", "Total USD", "Platform", "Date"]
    click.echo(tabulate(table_data, headers=headers, tablefmt="simple"))


@sale.command("remove")
@click.argument("sale_id")
@click.pass_context
def remove_sale(ctx, sale_id):
    """Remove a sale by ID."""
    portfolio = ctx.obj["portfolio"]
    if portfolio.remove_sale(sale_id):
        click.echo(f"Removed sale {sale_id}")
    else:
        click.echo(f"Sale {sale_id} not found")


# ============ Summary Command ============

@cli.command()
@click.pass_context
def summary(ctx):
    """Show portfolio summary."""
    portfolio = ctx.obj["portfolio"]
    summary = portfolio.get_summary()

    click.echo("\n=== PORTFOLIO SUMMARY ===\n")

    # Holdings by coin
    click.echo("HOLDINGS BY COIN:")
    if summary["holdings_by_coin"]:
        for coin, amount in sorted(summary["holdings_by_coin"].items()):
            click.echo(f"  {coin}: {amount:,.8f}".rstrip('0').rstrip('.'))
    else:
        click.echo("  (none)")

    # Sales by coin
    click.echo("\nSALES BY COIN:")
    if summary["sales_by_coin"]:
        for coin, amount in sorted(summary["sales_by_coin"].items()):
            click.echo(f"  {coin}: {amount:,.8f}".rstrip('0').rstrip('.'))
    else:
        click.echo("  (none)")

    # Loans by coin
    click.echo("\nLOANS BY COIN:")
    if summary["loans_by_coin"]:
        for coin, amount in sorted(summary["loans_by_coin"].items()):
            click.echo(f"  {coin}: {amount:,.8f}".rstrip('0').rstrip('.'))
    else:
        click.echo("  (none)")

    # Net holdings
    click.echo("\nNET HOLDINGS (Holdings - Sales - Loans):")
    if summary["net_by_coin"]:
        for coin, amount in sorted(summary["net_by_coin"].items()):
            prefix = "+" if amount > 0 else ""
            click.echo(f"  {coin}: {prefix}{amount:,.8f}".rstrip('0').rstrip('.'))
    else:
        click.echo("  (none)")

    click.echo(f"\n---------------------------")
    click.echo(f"Total Holdings: {summary['total_holdings_count']}")
    click.echo(f"Total Sales: {summary['total_sales_count']}")
    click.echo(f"Total Loans: {summary['total_loans_count']}")
    click.echo(f"Total Invested: ${summary['total_invested_usd']:,.2f}")
    click.echo(f"Total Sold: ${summary['total_sold_usd']:,.2f}")
    click.echo()


def main():
    cli(obj={})


if __name__ == "__main__":
    main()
