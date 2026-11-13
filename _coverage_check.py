import re
from pathlib import Path

spec = Path("api/openapi/WarpOpenAPIv2.yml").read_text(encoding="utf-8")
code = "\n".join(p.read_text(encoding="utf-8") for p in Path("alor").glob("*.go"))

lines = spec.splitlines()
ops = []
cur = None
in_paths = False
for i, line in enumerate(lines):
    if line.startswith("paths:"):
        in_paths = True
        continue
    if not in_paths:
        continue
    if line.startswith("components:"):
        break
    m = re.match(r"^  (/[^:]+):", line)
    if m:
        cur = m.group(1)
        continue
    m = re.match(r"^    (get|post|put|delete|patch):", line, re.I)
    if m and cur:
        ops.append((m.group(1).upper(), cur))

# Map path patterns to method name hints present in code as string literals
covered = []
missing = []
for method, path in ops:
    # normalize path template to searchable fragments
    # look for unique static segments in go source
    if path == "/refresh":
        ok = "/refresh" in code and "refreshAccessToken" in code
    else:
        # build regex-ish static parts
        static = re.sub(r"\{[^}]+\}", "", path)
        # find a distinctive substring from the path in code
        markers = [
            path,
            path.split("{")[0].rstrip("/"),
        ]
        # also try replacing templates with %s / %d patterns used in sprintf
        sprintf_like = re.sub(r"\{[^}]+\}", "%", path)
        ok = any(m and m in code for m in markers) or any(
            part and len(part) > 12 and part in code for part in path.split("/")
        )
        # stronger checks for specific endpoints
        specials = {
            "/md/v2/time": "GetTime",
            "/md/v2/Securities": "GetSecurities",
            "/md/v2/history": "GetHistory",
            "/md/v2/risk/rates": "GetRiskRates",
            "/md/v2/orderbooks/": "GetOrderBook",
            "/md/status/": "GetMarketStatus",
            "currencyPairs": "GetCurrencyPairs",
            "availableBoards": "GetAvailableBoards",
            "actualFuturesQuote": "GetActualFuturesQuote",
            "alltrades/history": "GetAllTradesHistory",
            "/alltrades": "GetAllTrades",
            "/quotes": "GetQuotes",
            "/orders/actions/market": "CreateMarketOrder",
            "/orders/actions/limit": "CreateLimitOrder",
            "/orders/actions/stopLimit": "CreateStopLimitOrder",
            "/orders/actions/stop": "CreateStopOrder",
            "orders/estimate/all": "EstimateOrders",
            "orders/estimate": "EstimateOrder",
            "orders/all": "DeleteAllOrders",
            "orderGroups": "OrderGroup",
            "stopLossLimit": "StopLossLimit",
            "takeProfitLimit": "TakeProfitLimit",
            "stopLoss": "StopLoss",
            "takeProfit": "TakeProfit",
            "fortsrisk": "GetFortsRisk",
            "/risk": "GetRisk",
            "/summary": "GetSummary",
            "/money": "GetMoney",
            "/positions": "GetPosition",
            "history/trades": "History",
            "stoporders": "GetStopOrder",
            "/Clients/": "GetOrder",
        }
        for marker, name in specials.items():
            if marker in path and name in code:
                ok = True
                break
        # DELETE single order commandapi
        if path.endswith("/client/orders/{orderId}") and "DeleteOrder" in code:
            ok = True
        if "warptrans/{tradeServerCode}/v2/client/orders/{orderId}" in path and "DeleteWarpOrder" in code:
            ok = True
    (covered if ok else missing).append(f"{method:6} {path}")

print(f"REST OpenAPI: {len(ops)}; covered~{len(covered)}; missing~{len(missing)}")
if missing:
    print("MISSING REST:")
    for m in missing:
        print(" ", m)

# WS subscription opcodes
idx = spec.find("field_WSOperationCodeSubscription:")
chunk = spec[idx : idx + 1200]
enums = re.findall(r"^      - (.+)$", chunk, re.M)
print("\nWS /ws opcodes:")
for e in enums:
    e = e.strip()
    present = e in code
    print(f"  [{'OK' if present else 'NO'}] {e}")

idx = spec.find("field_WSOperationCodeCommand:")
chunk = spec[idx : idx + 900]
enums = re.findall(r"^      - (.+)$", chunk, re.M)
print("\nWS /cws opcodes:")
for e in enums:
    e = e.strip()
    present = e in code
    print(f"  [{'OK' if present else 'NO'}] {e}")
