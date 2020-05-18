var searchParam = location.search;
var info;//包含如下key {id,room,username,live_url,ws_key,headImage}
var w; //websocket对象
var msgPanel;//消息面板
var members = {};//成员列表
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

/**
 * 显示错误
 */
function showError() {
    document.getElementById("error").style.display = "block";
    document.getElementById("room").style.display = "none";
}

/**
 * 显示房间
 */
function showRoom() {
    document.getElementById("room").style.display = "flex";
    document.getElementById("error").style.display = "none";
}

/**
 * 连接
 */
function connect() {
    w = new WebSocket("ws://" + window.location.host + "/ws?auth=" + encodeURIComponent(info.ws_key));
    w.onopen = function () {
        console.log("已连接");
        getMembers();
        flv_load();
    };

    w.onclose = function () {
        console.log("断开连接");
        flv_destroy();
    };
    w.onmessage = function (message) {
        var data = JSON.parse(message.data)
        resolveMsg(data)
    };
}

/**
 * 获取
 */
function getMembers() {
    axios.get('/members?room=' + info.room)
        .then(function (response) {
            let data = response.data;
            if (data.code == 0) {
                members = data.data;
                for (var k in members) {
                    createMember(k, members[k]);
                }
                showTotalMemberNum();
            } else {
                members = {}
            }
        }).catch(function (error) {
        members = {}
    });
}

/**
 * 消息处理器
 * @param data
 */
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
    actionHandler(data.id, data.username, data.action)
}

function actionHandler(id, name, action) {
    if (action == "leave") {
        delete members[id];
        document.getElementById(id).remove();
    } else if (action == "join") {
        if (!members[id]) {
            members[id] = name;
            createMember(id, name);
        }
    }
    showTotalMemberNum();
}

function showTotalMemberNum() {
    document.querySelector(".total-num").innerText = document.querySelectorAll(".member-item").length;
}

function createMember(id, username) {
    if (document.getElementById(id) != null) {
        return;
    }
    var div = document.createElement("div");
    div.setAttribute("id", id);
    div.classList.add("member-item");
    var image = document.createElement("img");
    image.setAttribute("src", getDefaultHeadImage(username.substring(username.length - 1)));
    div.appendChild(image);
    var span = document.createElement("span");
    span.innerText = username;
    div.appendChild(span);
    document.querySelector(".member-items").appendChild(div);
}

/**
 * 获取默认头像
 * @param txt
 * @returns {string}
 */
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

/**
 * 其他用户发言展示
 * @param data
 */
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

/**
 * 显示自己发言内容
 * @param data
 */
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

/**
 * 系统消息展示
 * @param msg
 */
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

/**
 * 发送消息
 */
function send() {
    var text = document.getElementById('sayText').value;
    if (text == "") {
        return
    }
    w.send(text);
    document.getElementById('sayText').value = "";
    document.getElementById('sayText').focus();
}

var flvPlayer;

/**
 * 打开直播画面
 * @param hasAudio
 */
function flv_load(hasAudio = true) {
    try {
        if (flvjs.isSupported()) {
            flv_destroy();
            var videoElement = document.getElementById('videoElement');
            flvPlayer = flvjs.createPlayer({
                type: 'flv',
                hasAudio: hasAudio,
                isLive: true,
                url: info.live_url
            });
            flvPlayer.attachMediaElement(videoElement);
            flvPlayer.on(flvjs.Events.METADATA_ARRIVED, function (e) {
                if ((!e.audiosamplerate || e.audiosamplerate === 0) && flvPlayer._mediaDataSource.hasAudio) {
                    flv_load(false);
                }
            })
            flvPlayer.load(); //加载
        }
    } catch (e) {
        console.error("打开直播失败", e)
    }
}

/**
 * 关闭直播画面
 */
function flv_destroy() {
    if (typeof flvPlayer !== "undefined" && flvPlayer != null) {
        flvPlayer.pause();
        flvPlayer.unload();
        flvPlayer.detachMediaElement();
        flvPlayer.destroy();
        flvPlayer = null;
    }
}

function keySend(event) {
    if (event.ctrlKey && event.key == "Enter") {
        send();
    }
}