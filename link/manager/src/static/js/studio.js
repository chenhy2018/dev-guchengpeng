var studioCategory = ""
var studioList = [];
var studioIndex = 0;
var studioRows = 20;
var studioTotal = 0;
var studioLike = "";

function fnGetEntry(uuid) {
    var ret = null;
    $.each(studioList, function(k, v){
        if (uuid == v.uuid) {
            ret = studioList[k];
            return false;
        }
    });
    return ret;
}

function fnFmtTime(seconds) {
    var date = new Date(seconds * 1000);
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
    var fmtTime = y + '/' + m + '/' + d + ' ' + h + ':' + min + ':' + s;
    return fmtTime;
}

function fnRefresh() {
    // render html
    var html = "";
    $.each(studioList, function(k, v){
        html += "<tr>";
        html += $("#tpl-item").tmpl({
            uuid: v.uuid,
            username: v.user,
            comment: v.comment,
            status: v.status,
            regtime: fnFmtTime(v.date),
            expire: fnFmtTime(v.expire)
        }).html();
        html += "</tr>";
    });
    $("#usr-list").html(html);

    // bind click events

    // edit button
    $(".edit-btn").click(function(){
        var item = fnGetEntry($(this).attr("uuid"));
        if (item != null) {
            fnShowDiag({
                uuid: item.uuid,
                title: "uuid: " + item.uuid,
                username: item.user,
                comment: item.comment,
                expire: fnFmtTime(item.expire)
            });
        } else {
            alert("没有" + $(this).attr("uuid") + "这个ID");
        }
    });

    // delete button
    $(".del-btn").click(function(){
        if (confirm("确定删除该配置？") == true) {
            $.ajax({
                type: 'POST',
                contentType: "application/json; charset=utf-8",
                url: '/api/studio',
                data: JSON.stringify({
                    "type":"studio",
                    "action":"delete",
                    "args":{
                        "uuid": $(this).attr("uuid")
                    }
                }),
                success: function (msg, status) {
                    fnDisplay();
                },
                error: function (xhr, desc, err) {
                    alert(xhr.responseText);
                }
            });
        }
    }); //.del-btn

    $(".grand-btn").click(function(){
        var item = fnGetEntry($(this).attr("uuid"));
        var status = ""
        if (item.status == "ok")
            status = ""
        else
            status = "ok"
        $.ajax({
            type: 'POST',
            contentType: "application/json; charset=utf-8",
            url: '/api/studio',
            data: JSON.stringify({
                "type":"studio",
                "action":"privilege",
                "args":{
                    "uuid": $(this).attr("uuid"),
                    "status": status
                }
            }),
            success: function (msg, status) {
                fnDisplay();
            },
            error: function (xhr, desc, err) {
                alert(xhr.responseText);
            }
        });
    }); //.grand-btn

    // paging
    html = "";
    for (i = 0; i < Math.ceil(studioTotal / studioRows); i++) {
        var active = "0";
        if (i == studioIndex) {
            active = "1";
        }
        html += $("#tpl-paging").tmpl({
            index : i + 1,
            active: active
        }).html();
    }
    $("#paging").html(html);

    // add empty tips
    if (studioList.length == 0) {
        $("#usr-list").html("<tr><td>当前没有配置信息</td><td></td><td></td><td></td><td></td><td></td><td></td></tr>");
    }
}

function fnDisplay() {
    var data = {
        "type": "studio",
        "action": "display",
        "args": {
            "index": studioIndex.toString(),
            "rows": studioRows.toString()
        }
    };
    if (studioLike != "") {
        data.args.category = studioCategory.toString();
        data.args.like = studioLike.toString();
    }

    // send req
    $.ajax({
        type:"POST",
        contentType: "application/json; charset=utf-8",
        url: '/api/studio',
        data: JSON.stringify(data),
        success: function (json, status) {
            json = JSON.parse(json)
            studioList = json.args.config;
            studioTotal = json.args.total;
            fnRefresh();
        }
    });
}

function fnShowDiag(obj) {
    $("#studio-diag").modal("show");
    var html = $("#tpl-edit").tmpl(obj).html();
    $("#studio-diag .studio-content").html(html);

    // save config
    $("#btn-save").click(function(){
        // disable button
        $(this).addClass('disabled');
        $(this).prop('disabled', true);

        var ajax = {
            type: 'POST',
            contentType: "application/json; charset=utf-8",
            url: '/api/studio'
        };

        ajax.error = function (xhr, desc, err) {
            $("#btn-save").removeClass('disabled');
            $("#btn-save").prop('disabled', false);
            alert(xhr.responseText);
            $("#studio-diag").modal("hide");
            fnDisplay();
        };
        ajax.success = function(msg, status) {
            $("#studio-diag").modal("hide");
            fnDisplay();
        }

        // yyyy/mm/dd hh:mm:ss to timestamp
        var fnGetTs = function(str) {
            var v = str.split(" ");
            if (v.length == 2) {
                var y = v[0].split("/")[0];
                var mon = v[0].split("/")[1];
                var d = v[0].split("/")[2];
                var h = v[1].split(":")[0];
                var m = v[1].split(":")[1];
                var r = new Date(y, mon - 1, d, h, m).getTime();
                return (isNaN(r) ? "0" : r / 1000);
            }
            return "0";
        }

        // build up json
        ajax.data =  JSON.stringify({
            "type":"studio",
            "action":"modify",
            "args": {
                "uuid": obj.uuid,
                "user": $("#username").val(),
                "comment": $("#comment").val(),
                "expire": fnGetTs($("#expire").val()).toString()
            }
        });

        // send add / modify
        $.ajax(ajax);
    });
}

function fnSearch() {
    studioCategory = $("#category").children("option:selected").val();
    studioLike = $("#search input").val();
    studioIndex = 0;
    fnDisplay();
}

function fnDirectTo(page) {
    $("#paging li").removeClass("active");
    studioIndex = page - 1;
    fnDisplay();
}

$(function(){
    fnDisplay();

    // press enter key
    $(document).keypress(function(e) {
        if (e.which == 13) {
            fnSearch();
        }
    });
});
