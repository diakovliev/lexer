require('../wasm_exec.js');
import "./index.css"
import "xterm/css/xterm.css"
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';

// input buffer
var buffer = ""
// input prompt
const PS = "? "

// initialize terminal instance
var term = new Terminal({
    cursorBlink: true,
    cursorStyle: 'underline',
    cursorInactiveStyle: "block",
    convertEol: true,
    altClickMovesCursor: false,
});
const fitAddon = new FitAddon();
term.loadAddon(fitAddon);
term.open(document.getElementById('terminal'));
fitAddon.fit();

if (!WebAssembly) {
    throw Error("WebAssembly is not supported in your browser")
}
    
const go = new Go();
WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then((result) => {
    go.run(result.instance);
    term.write(wasmWelcome());
    term.write(PS);
    term.focus();
    term.onKey((e) => {
        const ev = e.domEvent;
        const printable = !ev.altKey && !ev.ctrlKey && !ev.metaKey;
        // console.log(ev)
        if (ev.key === "Enter") {
            term.write("\n")
            term.write(wasmEvaluate(buffer))
            term.write(PS)
            buffer = ""
        } else if (ev.key === "Backspace") {
            if (buffer.length > 0) {
                term.write("\b \b");
                buffer = buffer.slice(0, -1)
            }
        } else if (ev.key.startsWith("Arrow")) {
            // ignore
        } else if (ev.key.startsWith("Page")) {
            // ignore
        } else if (ev.key === "Home" || ev.key === "End") {
            // ignore
        } else if (printable) {
            term.write(e.key);
            buffer += e.key
        }
    });
});
