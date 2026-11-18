"""Generate a standalone API reference from the generated Swagger document."""

import json
from html import escape
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
spec = json.loads((ROOT / "docs/swagger.json").read_text(encoding="utf-8"))
methods = {"get", "post", "put", "patch", "delete", "options", "head"}
endpoint_sections = []
for path, path_item in spec.get("paths", {}).items():
    for method, operation in path_item.items():
        if method not in methods:
            continue
        parameters = operation.get("parameters", [])
        parameter_rows = "".join(
            "<tr><td>{}</td><td>{}</td><td>{}</td><td>{}</td></tr>".format(
                escape(parameter.get("name", "")),
                escape(parameter.get("in", "")),
                "yes" if parameter.get("required") else "no",
                escape(parameter.get("description", "")),
            )
            for parameter in parameters
        )
        parameters_html = (
            "<table><thead><tr><th>Name</th><th>Location</th><th>Required</th><th>Description</th></tr></thead>"
            f"<tbody>{parameter_rows}</tbody></table>"
            if parameters
            else "<p class=muted>No parameters</p>"
        )
        responses = "".join(
            "<li><strong>{}</strong> {}</li>".format(
                escape(str(status)), escape(response.get("description", ""))
            )
            for status, response in operation.get("responses", {}).items()
        )
        endpoint_sections.append(
            f"""<section class="endpoint"><div class="endpoint-title"><span class="method {escape(method)}">{escape(method.upper())}</span><code>{escape(path)}</code></div>
<h3>{escape(operation.get("summary", ""))}</h3>{parameters_html}<h4>Responses</h4><ul>{responses}</ul></section>"""
        )
schema_sections = []
for name, schema in spec.get("definitions", {}).items():
    properties = "".join(
        "<tr><td>{}</td><td>{}</td><td>{}</td></tr>".format(
            escape(prop), escape(str(details.get("type", "object"))), escape(str(details.get("description", "")))
        )
        for prop, details in schema.get("properties", {}).items()
    )
    schema_sections.append(
        f"<section class=\"schema\"><h3>{escape(name)}</h3><table><thead><tr><th>Property</th><th>Type</th><th>Description</th></tr></thead><tbody>{properties}</tbody></table></section>"
    )
page = f"""<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>UptimeApp API Reference</title>
  <style>
    :root {{ color-scheme: light; font: 15px/1.5 Arial, sans-serif; color: #263238; background: #fff; }}
    body {{ margin: 0; }} header {{ padding: 32px 8%; color: #fff; background: #263238; }} main {{ max-width: 1100px; margin: 0 auto; padding: 24px; }}
    h1 {{ margin: 0 0 8px; }} h2 {{ margin-top: 40px; border-bottom: 2px solid #eee; padding-bottom: 8px; }} h3 {{ margin: 14px 0; }}
    .endpoint, .schema {{ border: 1px solid #e1e5e8; border-radius: 6px; margin: 18px 0; padding: 18px; }} .endpoint-title {{ display:flex; align-items:center; gap:12px; font-size: 18px; }}
    .method {{ border-radius: 4px; color:#fff; font-size:12px; font-weight:bold; padding:3px 8px; }} .get {{ background:#1683b5; }} .post {{ background:#49a942; }} .patch {{ background:#c58a00; }}
    code {{ background:#f4f5f6; padding:2px 5px; }} table {{ border-collapse: collapse; width: 100%; }} th, td {{ border:1px solid #e1e5e8; padding:8px; text-align:left; }} th {{ background:#f7f8f9; }} .muted {{ color:#68737d; }}
  </style>
</head>
<body>
  <header><h1>{escape(spec.get("info", dict()).get("title", "API Reference"))}</h1><div>{escape(spec.get("info", dict()).get("description", ""))}</div><div>Base path: <code>{escape(spec.get("basePath", ""))}</code></div></header>
  <main><h2>Endpoints</h2>{"".join(endpoint_sections)}<h2>Schemas</h2>{"".join(schema_sections)}</main>
</body>
</html>
"""
(ROOT / "docs/redoc.html").write_text(page, encoding="utf-8")
