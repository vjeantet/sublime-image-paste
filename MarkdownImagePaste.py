"""Markdown Image Paste - paste a clipboard image into a markdown file.

On Cmd+V in a markdown view, if the clipboard holds an image, the image is
saved next to the current .md file (or in a configurable subdirectory) as
``<mdname>_<N>.<ext>`` and a ``![](relative/path)`` reference is inserted.
Otherwise the normal paste is performed.

Clipboard access is delegated to a bundled Go helper, selected per platform:
``bin/darwin/imgpaste`` (macOS) or ``bin/windows/imgpaste.exe`` (Windows).
"""

import os
import re
import subprocess

import sublime
import sublime_plugin

SETTINGS_FILE = "MarkdownImagePaste.sublime-settings"
HELPER_TIMEOUT = 10  # seconds
IMAGE_EXTS = ("png", "jpg", "jpeg", "gif", "tiff", "bmp")

PLATFORM = sublime.platform()  # "osx", "windows" or "linux"


def _helper_path():
    base = os.path.join(os.path.dirname(__file__), "bin")
    if PLATFORM == "windows":
        return os.path.join(base, "windows", "imgpaste.exe")
    if PLATFORM == "osx":
        return os.path.join(base, "darwin", "imgpaste")
    return os.path.join(base, "linux", "imgpaste")


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
    kwargs = {
        "stdout": subprocess.PIPE,
        "stderr": subprocess.PIPE,
        "timeout": HELPER_TIMEOUT,
    }
    if PLATFORM == "windows":
        # Don't flash a console window when launching the helper.
        kwargs["creationflags"] = 0x08000000  # CREATE_NO_WINDOW
    proc = subprocess.run([helper] + args, **kwargs)
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
        if not file_name:
            sublime.status_message(
                "PasteImage: enregistre d'abord le fichier markdown"
            )
            return

        try:
            code, ext = _run_helper(["detect"])
        except Exception as e:
            sublime.status_message("PasteImage: erreur helper ({})".format(e))
            self.view.run_command("paste")
            return

        # No image in clipboard -> normal paste.
        if code != 0 or not ext:
            self.view.run_command("paste")
            return

        md_dir = os.path.dirname(file_name)
        stem = os.path.splitext(os.path.basename(file_name))[0]

        settings = sublime.load_settings(SETTINGS_FILE)
        subdir = settings.get("image_subdir", "") or ""
        target_dir = os.path.join(md_dir, subdir) if subdir else md_dir

        try:
            os.makedirs(target_dir, exist_ok=True)
        except OSError as e:
            sublime.status_message("PasteImage: dossier illisible ({})".format(e))
            return

        index = _next_index(target_dir, stem)
        out_name = "{}_{:02d}.{}".format(stem, index, ext)
        out_path = os.path.join(target_dir, out_name)

        try:
            save_code, _ = _run_helper(["save", out_path])
        except Exception as e:
            sublime.status_message("PasteImage: erreur sauvegarde ({})".format(e))
            return

        if save_code != 0 or not os.path.exists(out_path):
            sublime.status_message("PasteImage: échec de l'enregistrement de l'image")
            return

        rel_path = os.path.relpath(out_path, md_dir).replace(os.sep, "/")
        reference = "![]({})".format(rel_path)
        self.view.run_command("insert", {"characters": reference})
        sublime.status_message("PasteImage: image enregistrée -> {}".format(rel_path))
