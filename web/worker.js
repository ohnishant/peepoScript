// Runs the peepoScript wasm module off the main thread so a runaway
// loop can be stopped by terminating the whole worker.
//
// Protocol (postMessage):
//   -> {type: "run", id, code}       : execute code in the shared env
//   <- {type: "result", id, result}  : result = {ok, output, value, errors}
//   -> {type: "keywords"}            : ask for the language keywords
//   <- {type: "keywords", words}     : keyword literals, sorted
//   <- {type: "boot-error", error}   : wasm failed to load

let ready = false;
const queue = [];

function handle(msg) {
  if (msg.type === "run") {
    self.postMessage({ type: "result", id: msg.id, result: self.peepoRun(msg.code) });
  } else if (msg.type === "keywords") {
    self.postMessage({ type: "keywords", words: self.peepoKeywords() });
  }
}

self.onmessage = (e) => {
  if (!ready) {
    queue.push(e.data);
    return;
  }
  handle(e.data);
};

(async () => {
  try {
    self.importScripts("wasm_exec.js");
    const bytes = await (await fetch("peepo.wasm")).arrayBuffer();
    const go = new Go();
    const { instance } = await WebAssembly.instantiate(bytes, go.importObject);
    // main() blocks on select{} forever, so this promise stays pending
    // while the module serves peepoRun/peepoKeywords. The two globals
    // are registered synchronously inside run(), hence no await here.
    const lifetime = go.run(instance);
    lifetime.catch((err) => self.postMessage({ type: "boot-error", error: String(err) }));
    ready = true;
    queue.splice(0).forEach(handle);
  } catch (err) {
    self.postMessage({ type: "boot-error", error: String(err) });
  }
})();
