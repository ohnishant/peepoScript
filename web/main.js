// peepoScript playground: editor with twitch-style emote completion,
// wasm interpreter in a worker, 7tv emotes from towdan's channel.

const RUN_TIMEOUT_MS = 5000;
const TOWDAN_TWITCH_ID = "76020462"; // towdan's channel owns the preferred emote set
const AC_MAX_ITEMS = 8;

const editor = document.getElementById("editor");
const output = document.getElementById("output");
const runBtn = document.getElementById("run-btn");
const clearBtn = document.getElementById("clear-btn");
const statusEl = document.getElementById("emote-status");

const acEl = document.createElement("div");
acEl.id = "autocomplete";
document.body.append(acEl);

let emotes = new Map(); // name -> image url
let keywords = [];

let runId = 0;
let pendingRun = null; // {id, timer}
let worker = spawnWorker();

function spawnWorker() {
  const w = new Worker("worker.js");
  w.onmessage = onWorkerMessage;
  w.onerror = (e) => {
    // A panic inside the wasm module kills this worker for good; any
    // pending run would hang forever otherwise.
    appendOutput(`worker crashed: ${e.message}`, "err");
    if (!pendingRun) return;
    clearTimeout(pendingRun.timer);
    pendingRun = null;
    worker.terminate();
    worker = spawnWorker();
    runBtn.disabled = false;
    appendOutput("interpreter restarted; saved bindings are gone.", "err");
  };
  w.postMessage({ type: "keywords" });
  return w;
}

function onWorkerMessage(e) {
  const msg = e.data;
  switch (msg.type) {
    case "keywords":
      keywords = msg.words;
      loadEmotes(keywords);
      break;
    case "result":
      if (pendingRun && pendingRun.id === msg.id) {
        clearTimeout(pendingRun.timer);
        pendingRun = null;
        runBtn.disabled = false;
        renderResult(msg.result);
      }
      break;
    case "boot-error":
      appendOutput(`failed to load peepo.wasm: ${msg.error}`, "err");
      break;
  }
}

// ---------------------------------------------------------------- emotes

async function loadEmotes(keywords) {
  // 1) vendored copies shipped with the page (works offline + reliable)
  try {
    const manifest = await (await fetch("emotes/manifest.json")).json();
    for (const [name, file] of Object.entries(manifest.emotes ?? {})) {
      emotes.set(name, `emotes/${file}`);
    }
  } catch {
    // no manifest means nothing vendored
  }

  // 2) live 7tv fill for anything still missing (channel + global search)
  const missing = keywords.filter((n) => !emotes.has(n));
  if (missing.length > 0) {
    try {
      const channel = await fetchChannelEmotes();
      for (const name of missing) {
        if (channel.has(name)) emotes.set(name, channel.get(name));
      }
      const stillMissing = missing.filter((n) => !emotes.has(n));
      await Promise.all(stillMissing.map(async (name) => {
        const url = await searchEmote(name);
        if (url) emotes.set(name, url);
      }));
    } catch {
      // channel search failed; keep whatever we already have
    }
  }

  for (const url of emotes.values()) new Image().src = url;

  const vendored = keywords.length - missing.length;
  const live = missing.filter((n) => emotes.has(n)).length;
  const gaps = keywords.length - vendored - live;
  statusEl.textContent = `${vendored + live}/${keywords.length} symbols rendered as emotes` +
    (gaps ? ` (${gaps} unavailable on 7tv)` : "");
  statusEl.classList.toggle("ready", gaps === 0);
  if (gaps > 0) statusEl.classList.add("error");

  normalize(); // re-render whatever is already in the editor
  runBtn.disabled = false;
}

function emoteUrl(emote) {
  return `https:${emote.data.host.url}/2x.webp`;
}

async function fetchChannelEmotes() {
  const res = await fetch(`https://7tv.io/v3/users/twitch/${TOWDAN_TWITCH_ID}`);
  if (!res.ok) throw new Error(`7tv returned ${res.status}`);
  const data = await res.json();
  const found = new Map();
  for (const emote of data.emote_set?.emotes ?? []) {
    found.set(emote.name, emoteUrl(emote));
  }
  return found;
}

async function searchEmote(name) {
  const query = `query{search{all(query:${JSON.stringify(name)},perPage:1){emotes{items{id defaultName}}}}}`;
  const res = await fetch("https://7tv.io/v4/gql", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ query }),
  });
  if (!res.ok) return null;
  const data = await res.json();
  const hit = data.data?.search?.all?.emotes?.items?.[0];
  if (!hit || hit.defaultName !== name) return null;
  return `https://cdn.7tv.app/emote/${hit.id}/2x.webp`;
}

// ------------------------------------------------------------- editor
//
// The editor keeps a single invariant: its DOM always shows the source
// with emote names swapped for images. Every mutation (typing, paste,
// drop, programmatic insert) funnels through normalize(), which reads
// the plain source out of the DOM, rebuilds it, and puts the caret back
// where it was - measured in source characters so images count as the
// length of the names they stand in for.

function makeEmoteImg(name) {
  const img = document.createElement("img");
  img.className = "emote";
  img.src = emotes.get(name);
  img.alt = name;
  img.title = name;
  img.dataset.name = name;
  return img;
}

// Plain-text source; images contribute their literal names.
function editorSource(root = editor) {
  let src = "";
  for (const node of root.childNodes) {
    if (node.nodeType === Node.TEXT_NODE) {
      src += node.data;
    } else if (node.nodeName === "BR") {
      src += "\n";
    } else if (node.nodeName === "IMG") {
      src += node.dataset.name;
    } else {
      src += editorSource(node);
      if (node.nodeName === "DIV" || node.nodeName === "P") src += "\n";
    }
  }
  return src.replace(/\u00a0/g, " ");
}

// Walks leaf pieces (text, br, img) with their source-char lengths.
function forEachPiece(node, fn) {
  for (const child of node.childNodes) {
    if (child.nodeType === Node.TEXT_NODE) {
      fn(child, child.data.length);
    } else if (child.nodeName === "BR") {
      fn(child, 1);
    } else if (child.nodeName === "IMG") {
      fn(child, (child.dataset.name || "").length);
    } else {
      forEachPiece(child, fn);
    }
  }
}

function getCaretOffset() {
  const sel = getSelection();
  if (!sel.rangeCount || !sel.isCollapsed || !editor.contains(sel.anchorNode)) {
    return null;
  }
  let found = false;
  let count = 0;
  forEachPiece(editor, (node, len) => {
    if (found) return;
    if (node === sel.anchorNode) {
      count += Math.min(sel.anchorOffset, len);
      found = true;
    } else {
      count += len;
    }
  });
  if (!found) {
    // Caret sits between block elements; every piece was counted, so
    // count already holds the total source length.
  }
  return count;
}

function setCaretOffset(offset) {
  const boundaries = []; // {at, make}
  const walk = (node, base) => {
    let pos = base;
    for (const child of node.childNodes) {
      if (child.nodeType === Node.TEXT_NODE) {
        for (let i = 0; i <= child.data.length; i++) {
          boundaries.push({ at: pos + i, make: () => {
            const r = document.createRange();
            r.setStart(child, i);
            r.collapse(true);
            return r;
          }});
        }
        pos += child.data.length;
      } else if (child.nodeName === "BR") {
        boundaries.push({ at: pos + 1, make: () => {
          const r = document.createRange();
          r.setStartAfter(child);
          r.collapse(true);
          return r;
        }});
        pos += 1;
      } else if (child.nodeName === "IMG") {
        const len = (child.dataset.name || "").length;
        boundaries.push({ at: pos + len, make: () => {
          const r = document.createRange();
          r.setStartAfter(child);
          r.collapse(true);
          return r;
        }});
        pos += len;
      } else {
        pos = walk(child, pos);
      }
    }
    return pos;
  };
  walk(editor, 0);

  let best = null;
  for (const b of boundaries) {
    if (b.at <= offset) best = b;
    else break;
  }
  if (!best) return;
  const range = best.make();
  const sel = getSelection();
  sel.removeAllRanges();
  sel.addRange(range);
  editor.focus();
}

// Rebuilds the DOM from source text, swapping exact emote-name words
// for their images and wrapping string literals in a colored span.
// Identifiers split on everything non-word-like, so
// `peepoChat "Wokege".` renders inside the string quotes too.
function buildEditorFragment(src) {
  const frag = document.createDocumentFragment();
  // Strings run quote to quote with no escapes and no newline stop,
  // mirroring the lexer's readString. Unterminated quotes stay lit to
  // the end of input because that is exactly what readString will
  // swallow as ILLEGAL.
  for (const part of src.split(/("[^"]*(?:"|$))/)) {
    if (!part) continue;
    if (part.startsWith('"')) {
      const span = document.createElement("span");
      span.className = "str";
      appendEmoteParts(span, part);
      frag.append(span);
    } else {
      appendEmoteParts(frag, part);
    }
  }
  return frag;
}

function appendEmoteParts(root, text) {
  for (const part of text.split(/([A-Za-z0-9_]+)/)) {
    if (!part) continue;
    if (emotes.has(part)) root.append(makeEmoteImg(part));
    else root.append(part);
  }
}

function setEditorSource(src, caretOffset) {
  editor.replaceChildren(buildEditorFragment(src));
  setCaretOffset(caretOffset ?? src.length);
  resetUndoHistory();
}

function normalize() {
  const off = getCaretOffset();
  const src = editorSource(editor);
  editor.replaceChildren(buildEditorFragment(src));
  if (off !== null) setCaretOffset(off);
  scheduleUndoCommit();
}

editor.addEventListener("input", () => {
  normalize();
  updateAutocomplete();
});

// ---------------------------------------------------------------- undo
//
// Rebuilding the DOM on every keystroke throws away the browser's
// native undo history, so the editor keeps its own. A commit lands
// UNDO_DEBOUNCE_MS after the last mutation and stores the *previous*
// known state, so one undo step rewinds a whole typed burst, not just
// the final character. Programmatic replaces (autocomplete accept,
// example load) go through setEditorSource and restart the stack.

const UNDO_DEBOUNCE_MS = 400;
const UNDO_MAX_DEPTH = 200;

let undoStack = [];
let redoStack = [];
let committed = null; // {source, caretOffset}: newest state outside the stacks
let undoTimer = null;

function currentState() {
  return {
    source: editorSource(editor),
    caretOffset: getCaretOffset(),
  };
}

// Moves `committed` forward to what is in the editor right now,
// pushing the stale value onto the undo stack when the text changed.
function commitUndoPoint() {
  const cur = currentState();
  if (!committed) {
    // No baseline yet; startup order guarantees one exists today,
    // but a reorder shouldn't turn into a TypeError here.
    committed = cur;
    return;
  }
  if (cur.source !== committed.source) {
    undoStack.push(committed);
    if (undoStack.length > UNDO_MAX_DEPTH) undoStack.shift();
    redoStack = [];
  }
  committed = cur;
}

function scheduleUndoCommit() {
  clearTimeout(undoTimer);
  undoTimer = setTimeout(() => {
    undoTimer = null;
    commitUndoPoint();
  }, UNDO_DEBOUNCE_MS);
}

// Ctrl+z must never race a pending commit, or the burst it belongs to
// would be skipped entirely.
function flushUndoCommit() {
  if (!undoTimer) return;
  clearTimeout(undoTimer);
  undoTimer = null;
  commitUndoPoint();
}

function resetUndoHistory() {
  clearTimeout(undoTimer);
  undoTimer = null;
  undoStack = [];
  redoStack = [];
  committed = currentState();
}

function applySnapshot(snap) {
  editor.replaceChildren(buildEditorFragment(snap.source));
  setCaretOffset(snap.caretOffset ?? snap.source.length);
  committed = snap;
  hideAutocomplete(); // the word before the caret may no longer exist
}

function undoEdit() {
  flushUndoCommit();
  const prev = undoStack.pop();
  if (!prev) return;
  redoStack.push(currentState());
  applySnapshot(prev);
}

function redoEdit() {
  flushUndoCommit();
  const next = redoStack.pop();
  if (!next) return;
  undoStack.push(currentState());
  applySnapshot(next);
}

// Native undo is dead here (normalize() rewrites the DOM it would
// target), so these combos are always swallowed, history or not.
editor.addEventListener("keydown", (e) => {
  if (!(e.ctrlKey || e.metaKey)) return;
  const key = e.key.toLowerCase();
  if (key === "z" && !e.shiftKey) {
    e.preventDefault();
    undoEdit();
  } else if ((key === "z" && e.shiftKey) || key === "y") {
    e.preventDefault();
    redoEdit();
  }
});

// Paste as plain text at the caret; the input event then normalizes.
editor.addEventListener("paste", (e) => {
  e.preventDefault();
  const text = e.clipboardData?.getData("text/plain") ?? "";
  if (text) document.execCommand("insertText", false, text);
});

editor.addEventListener("blur", () => hideAutocompleteSoon());

// -------------------------------------------------------- autocomplete

let acItems = [];
let acIndex = 0;
let acPrefix = "";
let acHideTimer = null;

function completionCandidates(prefix) {
  const lower = prefix.toLowerCase();
  return [...new Set([...emotes.keys(), ...keywords])]
    .filter((n) => n.length >= prefix.length &&
      n.slice(0, prefix.length).toLowerCase() === lower)
    .sort();
}

function acOpen() {
  return acItems.length > 0;
}

function hideAutocomplete() {
  clearTimeout(acHideTimer);
  acItems = [];
  acEl.style.display = "none";
}

// Blur should close the popup, but a click on a popup row must not -
// mousedown there cancels the deferred hide before blur settles.
function hideAutocompleteSoon() {
  clearTimeout(acHideTimer);
  acHideTimer = setTimeout(hideAutocomplete, 150);
}

function currentWordBeforeCaret() {
  const off = getCaretOffset();
  if (off === null) return null;
  const match = editorSource(editor).slice(0, off).match(/[A-Za-z0-9_]+$/);
  if (!match) return null;
  return { offset: off, word: match[0] };
}

function updateAutocomplete() {
  const hit = currentWordBeforeCaret();
  if (!hit) return hideAutocomplete();

  const candidates = completionCandidates(hit.word).filter((n) => n !== hit.word);
  if (!candidates.length) return hideAutocomplete();

  clearTimeout(acHideTimer);
  acPrefix = hit.word;
  acItems = candidates.slice(0, AC_MAX_ITEMS);
  acIndex = 0;
  renderAcItems();
  positionAc(hit.offset);
}

function renderAcItems() {
  acEl.replaceChildren(...acItems.map((name, i) => {
    const row = document.createElement("div");
    row.className = "ac-item" + (i === acIndex ? " selected" : "");
    row.dataset.name = name;

    const label = document.createElement("span");
    label.textContent = name;
    row.append(label);

    if (emotes.has(name)) {
      const img = document.createElement("img");
      img.src = emotes.get(name);
      img.alt = "";
      row.append(img);
    } else {
      const pad = document.createElement("span"); // keep rows aligned
      pad.className = "ac-placeholder";
      row.append(pad);
    }
    return row;
  }));
  acEl.style.display = "block";
  acEl.querySelector(".selected")?.scrollIntoView({ block: "nearest" });
}

function positionAc(offset) {
  const sel = getSelection();
  if (!sel.rangeCount) return hideAutocomplete();

  // A collapsed caret can report empty rects (e.g. right after a
  // trailing newline); fall back to the last visible glyph before it.
  const range = sel.getRangeAt(0);
  const probe = document.createRange();
  probe.selectNodeContents(editor);
  probe.setEnd(range.endContainer, range.endOffset);
  const seen = [...probe.getClientRects()].filter((r) => r.width || r.height);
  const rect = seen.length ? seen[seen.length - 1] : editor.getBoundingClientRect();

  acEl.style.visibility = "hidden";
  acEl.style.display = "block";
  const box = acEl.getBoundingClientRect();
  const x = Math.max(8, Math.min(rect.left, window.innerWidth - box.width - 8));
  let y = rect.bottom + 4;
  if (y + box.height > window.innerHeight - 8) y = Math.max(8, rect.top - box.height - 4);
  acEl.style.left = `${x}px`;
  acEl.style.top = `${y}px`;
  acEl.style.visibility = "visible";
  void offset; // kept for callers; the selection range already encodes it
}

function acceptAutocomplete(name) {
  const hit = currentWordBeforeCaret();
  if (!hit) return hideAutocomplete();
  const src = editorSource(editor);
  const start = hit.offset - hit.word.length;
  const next = src.slice(0, start) + name + " " + src.slice(hit.offset);
  hideAutocomplete();
  setEditorSource(next, start + name.length + 1);
}

acEl.addEventListener("mousedown", (e) => {
  e.preventDefault(); // keep editor focus and caret intact
  const row = e.target.closest(".ac-item");
  if (row) acceptAutocomplete(row.dataset.name);
});

editor.addEventListener("keydown", (e) => {
  if (acOpen()) {
    if (e.key === "ArrowDown" || e.key === "ArrowUp") {
      e.preventDefault();
      const delta = e.key === "ArrowDown" ? 1 : -1;
      acIndex = (acIndex + delta + acItems.length) % acItems.length;
      renderAcItems();
    } else if (e.key === "Tab" || e.key === "Enter") {
      e.preventDefault();
      acceptAutocomplete(acItems[acIndex]);
    } else if (e.key === "Escape") {
      e.preventDefault();
      hideAutocomplete();
    }
    return;
  }

  if (e.key === "Tab") {
    // No popup yet: complete directly when there is exactly one
    // candidate, otherwise open the popup for the word at the caret.
    e.preventDefault();
    const hit = currentWordBeforeCaret();
    if (!hit) return;
    const cands = completionCandidates(hit.word).filter((n) => n !== hit.word);
    if (cands.length === 1) acceptAutocomplete(cands[0]);
    else updateAutocomplete();
  }
});

document.addEventListener("keydown", (e) => {
  if (e.key === "Enter" && e.ctrlKey) {
    e.preventDefault();
    runCode();
  }
});

// ---------------------------------------------------------------- run

function runCode() {
  if (runBtn.disabled || pendingRun) return;
  const code = editorSource(editor);
  if (!code.trim()) return;

  output.textContent = "";
  runBtn.disabled = true;
  runId += 1;
  const id = runId;
  const timer = setTimeout(() => {
    if (!pendingRun || pendingRun.id !== id) return;
    pendingRun = null;
    worker.terminate();
    worker = spawnWorker();
    runBtn.disabled = false;
    appendOutput(
      "run exceeded 5s and was killed (infinite loop?). " +
      "the worker restarted, so saved bindings are gone.",
      "err"
    );
  }, RUN_TIMEOUT_MS);
  pendingRun = { id, timer };

  worker.postMessage({ type: "run", id, code });
}

function renderResult(result) {
  output.textContent = "";
  if (result.output) appendOutput(result.output.replace(/\n$/, ""));
  if (result.value) appendOutput(result.value);
  for (const err of result.errors ?? []) appendOutput(err, "err");
}

function appendOutput(text, cls) {
  const div = document.createElement("div");
  if (cls) div.className = cls;
  div.textContent = text;
  output.append(div);
  output.scrollTop = output.scrollHeight;
}

runBtn.addEventListener("click", runCode);
clearBtn.addEventListener("click", () => (output.textContent = ""));

// ------------------------------------------------------------ examples

const EXAMPLES = {
  countdown: [
    "PepoG countdown SadgeBusiness n Wokege",
    "    Hmmge Scoots n 0 Wokege",
    "        peepoChat NODDERS.",
    "    Bedge peepoShrug Wokege",
    "        peepoChat n.",
    "        countdown PepegaCredit n 1.",
    "    Bedge",
    "Bedge",
    "",
    "countdown 3.",
    'peepoChat "liftoff".',
  ].join("\n"),
  greet: [
    "PepoG greet SadgeBusiness name Wokege",
    '    peepoFriendship "hello" name.',
    "Bedge",
    "",
    'greet "peepo".',
    'peepoChat "Wokege" "NODDERS".',
  ].join("\n"),
  operators: [
    "peepoFriendship mitosis 2 3 4.",
    "PepegaCredit 10 peepoBye 20 5.",
    "Scoots mitosis 2 2 4.",
    'peepoMeasure "Wokege".',
  ].join("\n"),
};

const examplesSelect = document.getElementById("examples");

function loadExample(key) {
  setEditorSource(EXAMPLES[key] ?? EXAMPLES.countdown);
  output.textContent = "";
}

examplesSelect?.addEventListener("change", (e) => loadExample(e.target.value));

// default
loadExample("countdown");
