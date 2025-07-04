var char_width = 0.0;
var char_height = 0.0;

function UpdateFontDimensions() {
    const code = document.createElement("code");
    const samples = 64;
    const elements = Array(samples).fill("J")
    code.innerHTML = elements.join("<br>")
    document.body.appendChild(code);
    char_width = code.getBoundingClientRect().width;
    char_height = code.getBoundingClientRect().height/samples;
    document.body.removeChild(code);
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
        if (activeTab != null) {
            activeTab.style.paddingBottom = "5px";
            let b = activeTab.getBoundingClientRect()
            let columns = Math.floor(b.width/char_width)
            let rows = Math.floor(b.height/char_height)
            
            let bottomPadding = b.height - (rows*char_height)
            this.tabElement.style.paddingBottom = bottomPadding + "px";
            activeTab.style.paddingBottom = bottomPadding + "px";

            if (columns != this.columns || rows != this.rows) {
                this.columns = columns
                this.rows = rows
                console.log(rows, columns, char_width, char_height)
                this.socket.send(JSON.stringify({"type":"size", "columns": columns, "rows": rows}))
            }
        }
    }
}