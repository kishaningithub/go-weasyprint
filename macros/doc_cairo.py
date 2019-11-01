""" Import the docsstring of Context for cairocffi """

import inspect
import re
import subprocess

import cairocffi

from style_accessor import camel_case

OUT_PATH = "drawer.go"

HEADER = """package goweasyprint

// autogenerated from cairocffi.py 

type float = pr.Float

// Drawer is the backend doing the actual drawing
// operations
type Drawer interface {{
    {meths}
}}
"""


meths = inspect.getmembers(cairocffi.Context, inspect.isfunction)


def format_default(argspec):
    if not argspec.defaults:
        return None
    N = len(argspec.defaults)
    out = "//"
    for arg, default in zip(argspec.args[-N:], argspec.defaults):
        out += f" {arg} = {default}"
    return out


def format_signature(argspec, type_args: dict):
    names = argspec.args
    if names[0] == "self":
        names = names[1:]
    names = [name + " " + type_args.get(name, "interface{}") for name in names]
    return ", ".join(names)


def add_comments(lines: str):
    return "\n".join("// " + l for l in lines.splitlines())


RE_TYPE = re.compile(r":type (\w+): (\w+)")


def parse_type(doc: str):
    lines = doc.splitlines(True)
    out = ""
    d = {}
    for line in lines:
        match = RE_TYPE.search(line)
        if match:
            arg, type_ = match.group(1), match.group(2)
            if arg == "float":  # inversion in doc string
                arg, type_ = type_, arg
            d[arg] = type_
        else:
            out += line
    return out, d


out = ""
for _, f in meths:
    name = camel_case(f.__name__)
    doc = inspect.getdoc(f)
    args = inspect.getfullargspec(f)
    if doc:
        doc, type_args = parse_type(doc)
        doc = add_comments(doc)
        default = format_default(args)
        fmt_args = format_signature(args, type_args)
        sig = f"{name}({fmt_args})"
        out += "\n" + doc + "\n"
        if default:
            out += default + "\n"
        out += sig + "\n"

with open(OUT_PATH, "w") as f:
    f.write(HEADER.format(meths=out))

subprocess.run(["goimports", "-w", OUT_PATH])
