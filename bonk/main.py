from __future__ import annotations

from .tui.app import BonkApp


def main() -> None:
    app = BonkApp()
    app.run()
