class Keyboard {
    constructor(tabElement, keyFunc, enterPressed) {
        this.tabElement = tabElement
        this.keyFunc = keyFunc;
        this.enterPressed = enterPressed
        this.enable_listeners()
    }
    enable_listeners() {
        document.addEventListener('keydown', e => {
            console.log(this.tabElement.previousSibling.previousSibling.checked)
            let active = this.tabElement.parentNode.classList.contains("active") && this.tabElement.previousElementSibling.checked
            let keyPressed = false;
            if (active) {
                if (!e.ctrlKey) {
                    if (e.key == "Enter") {
                        this.enterPressed()
                        this.keyFunc("\n")
                        keyPressed = true;
                    }
                    else if (e.key == "Tab") {
                        this.keyFunc("\t")
                        keyPressed = true;
                    }
                    else if (e.key == "ArrowUp") {
                        this.keyFunc("\u001b[A")
                        keyPressed = true;
                    }
                    else if (e.key == "ArrowDown") {
                        this.keyFunc("\u001b[B")
                        keyPressed = true;
                    }
                    else if (e.key == "ArrowLeft") {
                        this.keyFunc("\u001b[D")
                        keyPressed = true;
                    }
                    else if (e.key == "ArrowRight") {
                        this.keyFunc("\u001b[C")
                        keyPressed = true;
                    }
                    else if (e.key.length === 1) {
                        buffor += e.key
                        this.keyFunc(e.key)
                        keyPressed = true;
                    }
                    else if (e.key == "Backspace") {
                        this.keyFunc("\b")
                        keyPressed = true;
                    }
                    else if (e.key == "Escape") {
                        this.keyFunc("\u001b")
                        keyPressed = true;
                    }
                } else {
                    if (e.key == "a") {
                        this.keyFunc("\u0001")
                        keyPressed = true;
                    } else if (e.key == "e") {
                        this.keyFunc("\u0005")
                        keyPressed = true;
                    } else if (e.key == "c") {
                        this.keyFunc("\u0003")
                        keyPressed = true;
                    } else if (e.key == "z") {
                        this.keyFunc("\u001a")
                        keyPressed = true;
                    }
                    else if (e.key == "C") {
                        var sel = window.getSelection();
                        if (sel.rangeCount > 0) {
                            var range = sel.getRangeAt(0);
                            var selectedText = range.toString();
                            navigator.clipboard.writeText(selectedText);     
                            keyPressed = true;
                        }
                    }
                    else if (e.key == "V") {
                        navigator.clipboard.readText()
                            .then(text => {
                                this.keyFunc(text)
                            })
                            .catch(err => {
                                console.error('Failed to read clipboard contents: ', err);
                            });
                        keyPressed = true;
                    }
                }
                if (keyPressed) {
                    e.preventDefault();
                }
            }
        });
    }
}
