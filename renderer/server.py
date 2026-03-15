from flask import Flask, request, send_file, jsonify
import subprocess
import tempfile
import os
import sys
import time
from datetime import datetime

app = Flask(__name__)

RESET = "\033[0m"
RED = "\033[31m"
GREEN = "\033[32m"
YELLOW = "\033[33m"
GRAY = "\033[90m"

def _log(level, color, msg):
    ts = GRAY + datetime.now().strftime("%Y/%m/%d %H:%M:%S") + RESET
    print(f"{ts} {color}[{level}]{RESET}  {msg}", file=sys.stderr, flush=True)

def info(msg):
    _log("INFO", GREEN, msg)

def warn(msg):
    _log("WARN", YELLOW, msg)

def error(msg):
    _log("ERROR", RED, msg)

TEMPLATE = r"""
\documentclass[preview,border=16pt,varwidth]{standalone}
\usepackage{amsmath}
\usepackage{amssymb}
\usepackage{xcolor}
\usepackage{fix-cm}
\begin{document}
\color{white}
\fontsize{28}{34}\selectfont
%s
\end{document}
"""

@app.route("/render", methods=["POST"])
def render():
    data = request.get_json()
    if not data or "latex" not in data:
        return jsonify({"error": "Missing 'latex' field in request body"}), 400

    latex = data["latex"]
    start = time.time()
    preview = latex[:80] + ("..." if len(latex) > 80 else "")
    info(f"Starting render: {preview}")

    with tempfile.TemporaryDirectory() as tmp:
        tex_path = os.path.join(tmp, "input.tex")
        pdf_path = os.path.join(tmp, "input.pdf")
        png_path = os.path.join(tmp, "out.png")

        with open(tex_path, "w") as f:
            f.write(TEMPLATE % latex)

        # pdflatex
        try:
            t0 = time.time()
            result = subprocess.run(
                ["pdflatex", "-no-shell-escape", "-interaction=nonstopmode", "input.tex"],
                cwd=tmp,
                capture_output=True,
                text=True,
                timeout=30,
            )
            info(f"pdflatex finished in {time.time() - t0:.2f}s (exit={result.returncode})")
        except subprocess.TimeoutExpired:
            error("pdflatex timed out after 30s")
            return jsonify({"error": "LaTeX compilation timed out"}), 408

        if result.returncode != 0 or not os.path.exists(pdf_path):
            error(f"pdflatex failed:\n{result.stdout}")
            if result.stderr:
                error(f"pdflatex stderr:\n{result.stderr}")

            errors = [line for line in result.stdout.splitlines() if line.startswith("!")]
            return jsonify({
                "error": "LaTeX compilation failed",
                "details": errors,
            }), 400

        # PDF to transparent PNG via ImageMagick
        try:
            t0 = time.time()
            result = subprocess.run(
                ["convert", "-density", "300", pdf_path, "-background", "none", "-flatten", png_path],
                capture_output=True,
                text=True,
            )
            info(f"convert finished in {time.time() - t0:.2f}s (exit={result.returncode})")
        except Exception as e:
            error(f"convert failed: {e}")
            return jsonify({"error": f"PDF to PNG conversion failed: {e}"}), 500

        if result.returncode != 0 or not os.path.exists(png_path):
            error(f"convert failed:\n{result.stderr}")
            return jsonify({"error": "PDF to PNG conversion failed", "details": result.stderr}), 500

        elapsed = time.time() - start
        info(f"Render complete in {elapsed:.2f}s")

        return send_file(png_path, mimetype="image/png")

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8080)
