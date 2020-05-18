var searchParam = location.search;
var info;//包含如下key {id,room,username,live_url,ws_key,headImage}
var w; //websocket对象
var msgPanel;//消息面板
axios.get('/auth' + searchParam)
    .then(function (response) {
        let data = response.data;
        if (data.code == 0) {
            showRoom();
            info = JSON.parse(data.data);
            info.headImage = getDefaultHeadImage(info.username.substring(info.username.length - 1));
            msgPanel = document.getElementsByClassName("show")[0];
            connect();
        } else {
            showError();
        }
    }).catch(function (error) {
    showError();
});

function showError() {
    document.getElementById("error").style.display = "block";
    document.getElementById("room").style.display = "none";
}

function showRoom() {
    document.getElementById("room").style.display = "flex";
    document.getElementById("error").style.display = "none";
}

function connect() {
    w = new WebSocket("ws://" + window.location.host + "/ws?auth=" + encodeURIComponent(info.ws_key));
    w.onopen = function () {
        console.log("已连接");
    };

    w.onclose = function () {
        console.log("断开连接");
    };
    w.onmessage = function (message) {
        var data = JSON.parse(message.data)
        resolveMsg(data)
    };
}

function resolveMsg(data) {
    if (data.type == "st") {
        sysMsg(data.message)
    } else if (data.type == "ut") {
        if (data.id == info.id) {
            selfMsg(data)
        } else {
            userMsg(data)
        }
    }
}


function getDefaultHeadImage(txt) {

    var canvas = document.getElementById("canvas");
    var w = canvas.width;
    var h = canvas.height;
    var context = canvas.getContext("2d");
    //背景色
    context.fillStyle = "#5a97ff";
    //draw background / rect on entire canvas
    context.fillRect(0, 0, w, h);
    // 设置字体
    context.font = "18px bold 黑体";
    // 设置颜色
    context.fillStyle = "#fff";
    // 设置水平对齐方式
    context.textAlign = "center";
    // 设置垂直对齐方式
    context.textBaseline = "middle";
    // 绘制文字（参数：要写的字，x坐标，y坐标）
    context.fillText(txt, 20, 20);
    return canvas.toDataURL('image/jpeg');
}

function userMsg(data) {
    var div = document.createElement("div");
    div.classList.add("user-info");
    var divT = document.createElement("div");
    divT.classList.add("other-info");

    var image = document.createElement("img");
    image.setAttribute("src", info.headImage);
    image.style.width = "40px";
    image.style.height = "40px";
    divT.appendChild(image);

    var div1 = document.createElement("div");
    var div2 = document.createElement("div");
    div2.classList.add("say-name");
    div2.innerText = data.username + "  " + data.time;
    div1.appendChild(div2);
    var spanMsg = document.createElement("div");
    spanMsg.classList.add("bubble-l");
    spanMsg.innerText = data.message;
    div1.appendChild(spanMsg);
    divT.appendChild(div1);

    div.appendChild(divT);
    msgPanel.appendChild(div);
    msgPanel.scrollTop = msgPanel.scrollHeight;
}

function selfMsg(data) {
    var div = document.createElement("div");
    div.classList.add("user-info");
    var divT = document.createElement("div");
    divT.classList.add("self-info");
    var div1 = document.createElement("div");
    var div2 = document.createElement("div");
    div2.classList.add("say-name");
    div2.innerText = data.time + "  " + data.username;
    div1.appendChild(div2);

    var div3 = document.createElement("div");
    var spanMsg = document.createElement("div");
    spanMsg.classList.add("bubble");
    spanMsg.innerText = data.message;
    div3.appendChild(spanMsg);
    div1.appendChild(div3);
    divT.appendChild(div1);
    var image = document.createElement("img");
    image.setAttribute("src", info.headImage);
    image.style.width = "40px";
    image.style.height = "40px";
    divT.appendChild(image);

    div.appendChild(divT);
    msgPanel.appendChild(div);
    msgPanel.scrollTop = msgPanel.scrollHeight;
}

function sysMsg(msg) {
    var div = document.createElement("div");
    div.classList.add("sys-info");
    var spanMsg = document.createElement("span");
    spanMsg.classList.add("sys-msg");
    spanMsg.innerText = msg;
    div.appendChild(spanMsg);
    msgPanel.appendChild(div);
    msgPanel.scrollTop = msgPanel.scrollHeight;
}

function send() {
    var text = document.getElementById('sayText').value;
    if (text == "") {
        return
    }
    w.send(text);
    document.getElementById('sayText').value = "";
    document.getElementById('sayText').focus();
}