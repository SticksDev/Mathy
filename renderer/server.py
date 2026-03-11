from flask import Flask, request, send_file, jsonify
import subprocess
import tempfile
import os

app = Flask(__name__)

TEMPLATE = r"""
\documentclass[preview,border=2pt]{standalone}
\usepackage{amsmath}
\usepackage{amssymb}
\begin{document}
\[
%s
\]
\end{document}
"""

@app.route("/render", methods=["POST"])
def render():
    data = request.get_json()
    if not data or "latex" not in data:
        return jsonify({"error": "Missing 'latex' field in request body"}), 400

    latex = data["latex"]

    with tempfile.TemporaryDirectory() as tmp:
        tex_path = os.path.join(tmp, "input.tex")
        pdf_path = os.path.join(tmp, "input.pdf")
        svg_path = os.path.join(tmp, "out.svg")
        log_path = os.path.join(tmp, "input.log")

        with open(tex_path, "w") as f:
            f.write(TEMPLATE % latex)

        try:
            result = subprocess.run(
                ["pdflatex", "-no-shell-escape", "-interaction=nonstopmode", "input.tex"],
                cwd=tmp,
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL,
                timeout=5,
            )
        except subprocess.TimeoutExpired:
            return jsonify({"error": "LaTeX compilation timed out"}), 408

        if result.returncode != 0 or not os.path.exists(pdf_path):
            log = ""
            if os.path.exists(log_path):
                with open(log_path, "r") as f:
                    log = f.read()
            return jsonify({"error": "LaTeX compilation failed", "log": log}), 400

        try:
            result = subprocess.run(
                ["pdf2svg", "input.pdf", "out.svg"],
                cwd=tmp,
                capture_output=True,
                text=True,
            )
        except Exception as e:
            return jsonify({"error": f"PDF to SVG conversion failed: {e}"}), 500

        if result.returncode != 0 or not os.path.exists(svg_path):
            return jsonify({"error": "PDF to SVG conversion failed", "details": result.stderr}), 500

        return send_file(svg_path, mimetype="image/svg+xml")

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8080)
