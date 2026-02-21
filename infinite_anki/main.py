from __future__ import annotations

from .tui.app import InfiniteAnkiApp


def main() -> None:
    app = InfiniteAnkiApp()
    app.run()
