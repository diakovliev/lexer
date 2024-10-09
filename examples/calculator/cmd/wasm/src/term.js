import "xterm/css/xterm.css"
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';

export class Term {
    constructor(parentElementID) {
        this.parentElementID = parentElementID
        this.PS = "? "
        this.buffer = ""
        this.term = new Terminal({
            cursorBlink: true,
            cursorStyle: 'underline',
            cursorInactiveStyle: "block",
            convertEol: true,
            altClickMovesCursor: false,
        });
        this.fitAddon = new FitAddon();
        this.term.loadAddon(this.fitAddon);
    }

    onKey(e) {
        const ev = e.domEvent;
        const printable = !ev.altKey && !ev.ctrlKey && !ev.metaKey;
        // console.log(ev)
        if (ev.key === "Enter") {
            this.term.write("\n")
            this.term.write(wasmEvaluate(this.buffer))
            this.term.write(this.PS)
            this.buffer = ""
        } else if (ev.key === "Backspace") {
            if (buffer.length > 0) {
                this.term.write("\b \b");
                this.buffer = this.buffer.slice(0, -1)
            }
        } else if (ev.key.startsWith("Arrow")) {
            // ignore
        } else if (ev.key.startsWith("Page")) {
            // ignore
        } else if (ev.key === "Home" || ev.key === "End") {
            // ignore
        } else if (printable) {
            this.term.write(e.key);
            this.buffer += e.key
        }    
    }

    run() {
        this.term.open(document.getElementById(this.parentElementID));
        this.fitAddon.fit();
        this.term.write(wasmWelcome());
        this.term.write(this.PS);
        this.term.focus();        
        this.term.onKey(this.onKey.bind(this))
    }
}
