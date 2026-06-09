"""Markdown Image Paste - paste a clipboard image into a markdown file.

On Cmd+V in a markdown view, if the clipboard holds an image, the image is
saved next to the current .md file (or in a configurable subdirectory) as
``<mdname>_<N>.<ext>`` and a ``![](relative/path)`` reference is inserted.
Otherwise the normal paste is performed.

macOS only for now. Clipboard access is delegated to the bundled Go helper
``bin/darwin/imgpaste``.
"""

import os
import re
import subprocess

import sublime
import sublime_plugin

SETTINGS_FILE = "MarkdownImagePaste.sublime-settings"
HELPER_TIMEOUT = 10  # seconds
IMAGE_EXTS = ("png", "jpg", "jpeg", "gif", "tiff")


def _helper_path():
    return os.path.join(os.path.dirname(__file__), "bin", "darwin", "imgpaste")


def _ensure_executable(path):
    try:
        mode = os.stat(path).st_mode
        if not (mode & 0o111):
            os.chmod(path, mode | 0o111)
    except OSError:
        pass


def _run_helper(args):
    """Run the helper with the given args. Returns (returncode, stdout-stripped)."""
    helper = _helper_path()
    _ensure_executable(helper)
    proc = subprocess.run(
        [helper] + args,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        timeout=HELPER_TIMEOUT,
    )
    return proc.returncode, proc.stdout.decode("utf-8", "replace").strip()


def _next_index(target_dir, stem):
    """Return (max existing numeric suffix for `stem`) + 1, or 1 if none."""
    pattern = re.compile(
        r"^" + re.escape(stem) + r"_(\d+)\.(?:" + "|".join(IMAGE_EXTS) + r")$",
        re.IGNORECASE,
    )
    highest = 0
    try:
        for name in os.listdir(target_dir):
            m = pattern.match(name)
            if m:
                highest = max(highest, int(m.group(1)))
    except OSError:
        pass
    return highest + 1


class MdPasteImageCommand(sublime_plugin.TextCommand):
    def run(self, edit):
        file_name = self.view.file_name()

        try:
            code, ext = _run_helper(["detect"])
        except Exception as e:
            sublime.status_message("PasteImage: helper error ({})".format(e))
            self.view.run_command("paste")
            return

        # No image in clipboard -> normal paste.
        if code != 0 or not ext:
            self.view.run_command("paste")
            return

        # Image in clipboard but file not saved yet: no directory to save the
        # image into, so ask the user to save first.
        if not file_name:
            sublime.status_message(
                "PasteImage: save the markdown file first"
            )
            return

        md_dir = os.path.dirname(file_name)
        stem = os.path.splitext(os.path.basename(file_name))[0]

        settings = sublime.load_settings(SETTINGS_FILE)
        subdir = settings.get("image_subdir", "") or ""
        target_dir = os.path.join(md_dir, subdir) if subdir else md_dir

        try:
            os.makedirs(target_dir, exist_ok=True)
        except OSError as e:
            sublime.status_message("PasteImage: cannot create folder ({})".format(e))
            return

        index = _next_index(target_dir, stem)
        out_name = "{}_{:02d}.{}".format(stem, index, ext)
        out_path = os.path.join(target_dir, out_name)

        try:
            save_code, _ = _run_helper(["save", out_path])
        except Exception as e:
            sublime.status_message("PasteImage: save error ({})".format(e))
            return

        if save_code != 0 or not os.path.exists(out_path):
            sublime.status_message("PasteImage: failed to save image")
            return

        rel_path = os.path.relpath(out_path, md_dir).replace(os.sep, "/")
        reference = "![]({})".format(rel_path)
        self.view.run_command("insert", {"characters": reference})
        sublime.status_message("PasteImage: image saved -> {}".format(rel_path))
