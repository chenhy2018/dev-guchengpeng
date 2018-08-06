var nodeList = [];

function fnGetNodes() {
    var data = {
        "type":"mgmt",
        "action":"lsnode",
        "args":{}
    };

    // send req
    $.ajax({
        type: 'POST',
        async: false,
        contentType: "application/json; charset=utf-8",
        url: '/api/action',
        data: JSON.stringify(data),
        success: function (json, status) {
            nodeList = json.args.result;
        },
        error: function (xhr, desc, err) {
            switch (xhr.status){
            case 403:
                alert("无授权查看该页面");
                window.location.href='/';
                break;
            default:
                alert("不能获取运行配置");
            }
        }
    });
}

function fnQuery(index)
{
    var data = {
        "type":"mgmt",
        "action":"audit",
        "args":{}
    };

    // choose index and rows
    var rows = 100;
    data.args.index = index.toString();
    data.args.rows = rows.toString();

    // choose node
    if ($("#nodelist").val() != "0_0") {
        data.args.disp_id = $("#nodelist").val();
    }

    // choose duration
    var fnGetTs = function(str) {
        var v = str.split(" ");
        if (v.length == 2) {
            var y = v[0].split("/")[0];
            var mon = v[0].split("/")[1];
            var d = v[0].split("/")[2];
            var h = v[1].split(":")[0];
            var m = v[1].split(":")[1];
            var r = new Date(y, mon - 1, d, h, m).getTime();
            return (isNaN(r) ? "" : r / 1000);
        }
        return "";
    }
    var start = fnGetTs($("#start").val());
    if (start != "") {
        data.args.start_time = start.toString();
    }
    var end = fnGetTs($("#end").val());
    if (end != "") {
        data.args.end_time = end.toString();
    }

    // choose tags
    $.each($("#tags").tagsinput('items'), function(k, v) {
        var kv = v.split(":");
        if (kv.length == 2) {
            if (data.args.context == null) {
                data.args.context = {};
            }
            data.args.context[kv[0]] = kv[1];
        }
    });

    // choose level mask
    var mask = $("#debug").attr("sw") << 5 | $("#info").attr("sw") << 4 |
        $("#notice").attr("sw") << 3 | $("#warn").attr("sw") << 2 | $("#error").attr("sw") << 1;
    data.args.mask = mask.toString();

    // choose keyword
    if ($("#keyword").val() != "") {
        data.args.like = $("#keyword").val();
    }

    // send req
    $.ajax({
        type: 'POST',
        async: false,
        contentType: "application/json; charset=utf-8",
        url: '/api/action',
        data: JSON.stringify(data),
        success: function (json, status) {
            fnShow(json.args.result);
            fnPaging(json.args.total, index, rows);
        },
        error: function (xhr, desc, err) {
            switch (xhr.status){
            case 403:
                alert("无授权查看该页面");
                window.location.href='/';
                break;
            default:
                alert(xhr.responseText);
            }
        }
    });
}

function fnShow(result)
{
    var html = ""
    $.each(result, function(k, v){
        html += $("#tpl-log").tmpl({
            "level": v.level,
            "message": fnFmtMsg(v)
        }).html();
    });
    $("#log-box").html(html);
}

function fnPaging(total, index, rows)
{
    var html = ""
    for (var i = 0; i < Math.ceil(total / rows); i++) {
        var active = "0";
        if (i == index) {
            active = "1";
        }
        html += $("#tpl-paging").tmpl({
            index: i,
            page: i + 1,
            active: active
        }).html();
    }
    $(".pagination").html(html);
}

function fnFmtMsg(obj)
{
    var logMsg = ""

    // time
    var date = new Date(obj.time / 1000000);
    var y = date.getFullYear();
    var m = date.getMonth() + 1;
    m = m < 10 ? ('0' + m) : m;
    var d = date.getDate();
    d = d< 10 ? ('0' + d) : d;
    var h = date.getHours();
    h = h < 10 ? ('0' + h) : h;
    var min = date.getMinutes();
    min = min < 10 ? ('0' + min) : min;
    var s = date.getSeconds();
    s = s < 10 ? ('0' + s) : s;
    var milli = date.getMilliseconds();
    milli = milli < 10 ? ('00' + milli) : (milli < 100 ? '0' + milli : milli);
    var fmtTime = y + '/' + m + '/' + d + ' ' + h + ':' + min + ':' + s + '.' + milli;

    // content
    var fmtMsg = obj.message;

    // file
    var fmtFile = '[' + obj.file + ':' + obj.line + ']';

    // context
    var fmtCtx = ""
    $.each(obj.context, function(k, v){
        fmtCtx += "[" + k + ":" + v + "]";
    });

    return fmtTime + ' ' + fmtFile + fmtCtx + ' ' + fmtMsg;
}

function fnLv(dom)
{
    if ($(dom).attr("sw") == "1") {
        $(dom).addClass("disabled");
        $(dom).attr("sw", "0");
    } else {
        $(dom).removeClass("disabled");
        $(dom).attr("sw", "1");
    }
}

$(function(){
    // get node list
    fnGetNodes();
    var html = '<option value="0_0">中心</option>';
    $.each(nodeList, function(k, v){
        html += '<option value="' + v.disp_id + '">' + v.desc + '</option>';
    });
    $("#nodelist").html(html);

    // get logs within last 1h
    fnQuery(0);
})
