var char_width = 0;
var char_height = 0

function UpdateFontDimensions() {
    const span = document.createElement("code");
    span.innerHTML = "x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x<br>x"
    span.style.width = "1ch";
    span.style.position = "fixed";

    span.style.fontFamily = "monospace";
    span.style.lineHeight = "1.2";

    document.body.appendChild(span);
    char_width = span.getBoundingClientRect().width;
    char_height = span.getBoundingClientRect().height/160;
    document.body.removeChild(span);
}

document.addEventListener("DOMContentLoaded", function(event) {
  UpdateFontDimensions();
});

class Terminal {

    constructor(session_id, tabElement, socket) {
        this.session_id = session_id;
        this.tabElement = tabElement
        this.element = document.getElementById(this.session_id);
        window.addEventListener("resize", this.UpdateSize.bind(this));
        this.lines = 0;
        this.chars = 0;

        this.socket = socket;

        this.keyboard = new Keyboard(this.tabElement, function(c){
            this.socket.send(JSON.stringify({"type":"keyboard", "keys": c}))
        }.bind(this),
        function() {
            const isScrolledToBottom = this.tabElement.scrollHeight - this.tabElement.clientHeight <= this.tabElement.scrollTop + 1
            if (!isScrolledToBottom) {
                this.tabElement.scrollTop = this.tabElement.scrollHeight - this.tabElement.clientHeight
            }
        }.bind(this));
        this.UpdateSize();
    }

    UpdateSize() {
        const tabInputs = this.tabElement.parentNode.getElementsByTagName('input');
        var activeTab = null
        for(var i = 0; i < tabInputs.length; i++){
            if (tabInputs[i].checked) {
                activeTab = tabInputs[i].nextElementSibling
            }
        }
        const margin = 10
        if (activeTab != null) {
            activeTab.style.height = ''
            let b = activeTab.getBoundingClientRect()
            console.log(b)

            let columns = Math.floor((b.width-margin)/char_width)
            let rows = Math.floor((b.height-margin)/char_height)
            
            if (columns != this.columns || rows != this.rows) {
                this.columns = columns
                this.rows = rows
                console.log(rows, columns, char_width, char_height)
                this.socket.send(JSON.stringify({"type":"size", "columns": columns, "rows": rows}))
            }
            activeTab.style.height = (rows*char_height+1).toString() + 'px'
        }
    }
}