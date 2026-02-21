from __future__ import annotations

import time

from textual.app import App, ComposeResult
from textual.containers import Container, Horizontal
from textual.widgets import Button, Footer, Header, Input, ListItem, ListView, Static

from ..db import open_db
from ..engine import due_skills, finish_review, pick_followup, start_review


SOCRATIC_CAP_SECONDS = 60


class InfiniteAnkiApp(App):
    CSS = """
    Screen { align: center middle; }
    #root { width: 100%; height: 100%; }
    #left { width: 40%; border: solid #2b2b2b; }
    #right { width: 60%; border: solid #2b2b2b; }
    .box { padding: 1; }
    #prompt { height: 10; overflow-y: auto; }
    #followup { height: 6; overflow-y: auto; }
    #log { height: 6; overflow-y: auto; }
    """

    BINDINGS = [
        ("q", "quit", "Quit"),
    ]

    def __init__(self):
        super().__init__()
        self.db = open_db()
        self.current = None  # ReviewSession
        self.started_ts = None
        self.followup = None
        self.failure_mode = None
        self.key_revealed = False

    def compose(self) -> ComposeResult:
        yield Header(show_clock=True)
        with Horizontal(id="root"):
            with Container(id="left", classes="box"):
                yield Static("Today (due skills)")
                yield ListView(id="skills")
                yield Button("Start", id="start", variant="primary")
            with Container(id="right", classes="box"):
                yield Static("Prompt")
                yield Static("", id="prompt")
                yield Static("Your answer")
                yield Input(placeholder="Explain your approach...", id="a1")
                yield Button("Submit", id="submit", variant="primary")
                yield Static("Coach follow-up")
                yield Static("", id="followup")
                yield Input(placeholder="Answer follow-up (optional)...", id="a2")
                with Horizontal():
                    yield Button("Again", id="rate1")
                    yield Button("Hard", id="rate2")
                    yield Button("Good", id="rate3")
                    yield Button("Easy", id="rate4")
                yield Static("", id="log")
        yield Footer()

    def on_mount(self) -> None:
        self.refresh_skills()
        self.set_focus(self.query_one("#skills", ListView))

    def refresh_skills(self) -> None:
        lv = self.query_one("#skills", ListView)
        lv.clear()
        skills = due_skills(self.db)
        for s in skills:
            lv.append(ListItem(Static(f"[{s.pattern}] {s.title}", expand=True), id=s.id))

    def selected_skill_id(self) -> str | None:
        lv = self.query_one("#skills", ListView)
        item = lv.highlighted_child
        return item.id if item else None

    def set_status(self, msg: str) -> None:
        w = self.query_one("#log", Static)
        w.update(msg)

    def action_start_review(self) -> None:
        sid = self.selected_skill_id()
        if not sid:
            self.set_status("Select a skill first.")
            return
        self.current = start_review(self.db, sid)
        self.started_ts = time.time()
        self.followup = None
        self.failure_mode = None
        self.key_revealed = False
        self.query_one("#prompt", Static).update(self.current.prompt)
        self.query_one("#followup", Static).update("")
        self.query_one("#a1", Input).value = ""
        self.query_one("#a2", Input).value = ""
        self.set_status("Started.")
        self.set_focus(self.query_one("#a1", Input))

    def on_button_pressed(self, event: Button.Pressed) -> None:
        bid = event.button.id
        if bid == "start":
            self.action_start_review()
            return
        if bid == "submit":
            self.handle_submit()
            return
        if bid in {"rate1", "rate2", "rate3", "rate4"}:
            rating = int(bid[-1])
            self.handle_finish(rating)
            return

    def handle_submit(self) -> None:
        if not self.current:
            self.set_status("Start a review first.")
            return
        a1 = self.query_one("#a1", Input).value.strip()
        if not a1:
            self.set_status("Type an answer.")
            return

        elapsed = time.time() - (self.started_ts or time.time())
        if elapsed > SOCRATIC_CAP_SECONDS and not self.key_revealed:
            # reveal key property only
            self.key_revealed = True
            self.query_one("#followup", Static).update(
                f"Key property: {self.current.key_property}\n\nTry again (you have one more shot)."
            )
            self.set_status("Key property revealed.")
            return

        q, failure = pick_followup(self.current.skill, a1)
        self.followup = q
        self.failure_mode = failure
        self.query_one("#followup", Static).update(q or "(no follow-up)")
        self.set_focus(self.query_one("#a2", Input))
        self.set_status("Follow-up asked. Rate when ready.")

    def handle_finish(self, rating: int) -> None:
        if not self.current:
            self.set_status("No active review.")
            return
        a2 = self.query_one("#a2", Input).value.strip() or None
        key_reveal = self.current.key_property if self.key_revealed else None
        next_due_at = finish_review(self.db, self.current.review_id, rating, answer2=a2, key_property_revealed=key_reveal)
        self.set_status(f"Done. Next due: {next_due_at}")
        self.refresh_skills()
        # keep prompt visible; user can start another.

"""
Keyboard shortcuts could be added next (enter to submit, 1-4 to rate).
"""
