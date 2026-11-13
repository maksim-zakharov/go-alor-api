#!/usr/bin/env python3
"""Генерирует Go-типы Simple из WarpOpenAPIv2.yml."""
from __future__ import annotations

import re
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parent
SPEC = ROOT / "api" / "openapi" / "WarpOpenAPIv2.yml"
OUT = ROOT / "alor" / "types_simple_gen.go"


def go_name(name: str) -> str:
    n = name
    # спец. префиксы для избежания коллизий object_ vs response_
    if n.startswith("object_"):
        n = "Obj" + n[len("object_") :]
    elif n.startswith("array_"):
        n = "Arr" + n[len("array_") :]
    elif n.startswith("response_"):
        n = n[len("response_") :]
    elif n.startswith("ws_request_sub_"):
        n = "WSReq" + n[len("ws_request_sub_") :]
    for suf in ("_Simple", "_Slim", "_Heavy"):
        if n.endswith(suf):
            n = n[: -len(suf)]
    parts = re.split(r"[^A-Za-z0-9]+", n)
    out = "".join(p[:1].upper() + p[1:] for p in parts if p)
    if out and out[0].isdigit():
        out = "N" + out
    # acronyms
    out = (
        out.replace("Id", "ID")
        .replace("Utc", "UTC")
        .replace("Isin", "ISIN")
        .replace("Cfi", "CFI")
        .replace("Url", "URL")
        .replace("Http", "HTTP")
        .replace("Ws", "WS")
        .replace("Api", "API")
        .replace("Cmd", "Cmd")
    )
    return out or "T"


def field_name(prop: str) -> str:
    parts = re.split(r"[^A-Za-z0-9]+", prop)
    out = "".join(p[:1].upper() + p[1:] for p in parts if p)
    if out.lower() in {
        "type",
        "package",
        "func",
        "var",
        "const",
        "map",
        "chan",
        "go",
        "range",
        "import",
        "interface",
        "struct",
        "return",
        "select",
        "case",
        "default",
        "switch",
        "if",
        "else",
        "for",
        "break",
        "continue",
    }:
        out = "Field" + out
    return (
        out.replace("Id", "ID")
        .replace("Utc", "UTC")
        .replace("Isin", "ISIN")
        .replace("Cfi", "CFI")
        .replace("Url", "URL")
        .replace("Http", "HTTP")
        .replace("Ws", "WS")
        .replace("Api", "API")
    )


def resolve_ref_name(ref: str) -> str:
    return ref.rsplit("/", 1)[-1]


def leaf_type(schema: dict) -> str:
    if not schema:
        return "any"
    t = schema.get("type")
    fmt = schema.get("format", "")
    if t == "integer":
        return "int32" if fmt == "int32" else "int64"
    if t == "number":
        return "float64"
    if t == "boolean":
        return "bool"
    if t == "string":
        return "string"
    if t == "array":
        return "[]any"
    if t == "object":
        return "map[string]any"
    if "enum" in schema:
        return "string"
    return "any"


def field_go_type(schemas: dict, schema: dict) -> tuple[str, bool]:
    if schema is None:
        return "any", False
    if "$ref" in schema:
        ref = resolve_ref_name(schema["$ref"])
        target = schemas.get(ref, {}) or {}
        if ref.startswith(("field_", "flag_", "path_", "query_")):
            return leaf_type(target), bool(target.get("nullable"))
        return go_name(ref), bool(target.get("nullable"))
    if "allOf" in schema:
        for part in schema["allOf"]:
            if isinstance(part, dict) and ("$ref" in part or "type" in part or "properties" in part):
                return field_go_type(schemas, part)
        return "any", False
    t = schema.get("type")
    if t == "array":
        it, _ = field_go_type(schemas, schema.get("items", {}))
        return "[]" + it, False
    if t == "object":
        return "map[string]any", False
    return leaf_type(schema), bool(schema.get("nullable"))


def should_emit(name: str) -> bool:
    if name.endswith(("_Heavy", "_Slim")):
        return False
    if name.startswith(("field_", "flag_", "path_", "query_")):
        return False
    if "Code400" in name or "CreationError" in name:
        return False
    if name.startswith(("response_", "object_", "array_")):
        return True
    if name == "ws_request_sub_BaseObject":
        return True
    return False


def emit(schemas: dict, name: str, schema: dict) -> str | None:
    gname = go_name(name)
    desc = str(schema.get("description", name)).splitlines()[0][:120]
    t = schema.get("type")
    if t == "array":
        it, _ = field_go_type(schemas, schema.get("items", {}))
        return f"// {gname} — {desc}\ntype {gname} = []{it}\n"
    props = schema.get("properties") or {}
    if t != "object" and not props:
        return None
    lines = [f"// {gname} — {desc}", f"type {gname} struct {{"]
    for prop, pschema in props.items():
        gt, nullable = field_go_type(schemas, pschema)
        fname = field_name(prop)
        if nullable and not gt.startswith(("[]", "map[", "*")):
            gt = "*" + gt
        lines.append(f'\t{fname} {gt} `json:"{prop},omitempty"`')
    lines.append("}")
    return "\n".join(lines) + "\n"


def main() -> None:
    data = yaml.safe_load(SPEC.read_text(encoding="utf-8"))
    schemas = data["components"]["schemas"]
    chunks = [
        "// Code generated by _gen_types.py from WarpOpenAPIv2.yml. DO NOT EDIT.",
        "package alor",
        "",
    ]
    # object_/array_ first, then response_ (response names stay short)
    order = sorted(
        (n for n in schemas if should_emit(n)),
        key=lambda n: (0 if n.startswith("object_") else 1 if n.startswith("array_") else 2, n),
    )
    seen_go: dict[str, str] = {}
    for name in order:
        code = emit(schemas, name, schemas[name])
        if not code:
            continue
        gname = go_name(name)
        if gname in seen_go:
            # rename collision: prefix with source kind
            print(f"skip duplicate Go name {gname}: {name} (kept {seen_go[gname]})")
            continue
        seen_go[gname] = name
        chunks.append(code)
    OUT.write_text("\n".join(chunks), encoding="utf-8")
    print(f"Wrote {OUT} ({len(seen_go)} types)")


if __name__ == "__main__":
    main()
