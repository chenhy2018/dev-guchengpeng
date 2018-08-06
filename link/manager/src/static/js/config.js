var conflist = [];
var confIndex = 0;
var confRows = 20;
var confTotal = 0;
var confLike = "";
var confNodes = [];

function fnGetEntry(job) {
    var ret = null;
    $.each(conflist, function(k, v){
        if (job == v.job) {
            ret = conflist[k];
            return false;
        }
    });
    return ret;
}

function fnFmtPolicy(str) {
    var obj = {};
    $.each(str.split("\n"), function(k, v){
        v = v.replace(/\r/g, "");
        if (!v.match(/=/)) {
            return true;
        }
        var pair = v.split("=");
        obj[pair[0]] = pair[1];
    });
    return obj;
}

function fnRefresh() {
    // render html
    var html = "";
    $.each(conflist, function(k, v){
        html += "<tr>";
        html += $("#tpl-item").tmpl({
            job: v.job,
            command: v.command
        }).html();
        html += "</tr>";
    });
    $("#conf-list").html(html);

    // bind click events

    // edit button
    $(".edit-btn").click(function(){
        var item = fnGetEntry($(this).attr("job"));
        if (item != null) {
            var toolString = item.command.split(" ")[0];
            var argString = item.command.replace(new RegExp("^" + toolString + "\\s*"), "");
            var policyString = "";
            $.each(item.policy, function(k,v){
                policyString += k + "=" + v + "\n";
            });
            fnShowDiag({
                mode: "edit",
                title: "编辑推流配置",
                tool:  toolString,
                job: item.job,
                args: argString,
                policy: policyString
            });
        } else {
            alert("没有" + $(this).attr("job") + "这个任务");
        }
    });

    // delete button
    $(".del-btn").click(function(){
        if (confirm("确定删除该配置？") == true) {
            $.ajax({
                type: 'POST',
                contentType: "application/json; charset=utf-8",
                url: '/api/action',
                data: JSON.stringify({
                    "type":"mgmt",
                    "action":"delete",
                    "args":{
                        "job": $(this).attr("job")
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

    html = "";
    for (i = 0; i < Math.ceil(confTotal / confRows); i++) {
        var active = "0";
        if (i == confIndex) {
            active = "1";
        }
        html += $("#tpl-paging").tmpl({
            index : i + 1,
            active: active
        }).html();
    }
    $("#paging").html(html);

    // add empty tips
    if (conflist.length == 0) {
        $("#conf-list").html("<tr><td>当前没有配置信息</td><td></td><td></td></tr>");
    }
}

function fnDisplay() {
    var data = {
        "type":"mgmt",
        "action":"display",
        "args":{
            "index":confIndex.toString(),
            "rows":confRows.toString()
        }

    };
    if (confLike != "") {
        data.args.like = confLike.toString();
    }

    // send req
    $.ajax({
        type: 'POST',
        contentType: "application/json; charset=utf-8",
        url: '/api/action',
        data: JSON.stringify(data),
        success: function (json, status) {
            conflist = json.args.config;
            confTotal = json.args.total;
            fnRefresh();
        }
    });
}

function fnShowDiag(obj) {
    $("#conf-diag").modal("show");
    var html = $("#tpl-conf").tmpl(obj).html();
    $("#conf-diag .conf-content").html(html);
    if ($('#conf-diag [data-toggle="select"]').length) {
        $('#conf-diag [data-toggle="select"]').select2();
    }

    // save config
    $("#btn-save").click(function(){
        // disable button
        $(this).addClass('disabled');
        $(this).prop('disabled', true);

        var ajax = {
            type: 'POST',
            contentType: "application/json; charset=utf-8",
            url: '/api/action'
        };

        ajax.error = function (xhr, desc, err) {
            $("#btn-save").removeClass('disabled');
            $("#btn-save").prop('disabled', false);
            alert(xhr.responseText);
            var json = xhr.responseText;
            switch (json.args.error) {
            case "not_effective":
                $("#conf-diag").modal("hide");
                fnDisplay();
            }
        };
        ajax.success = function(msg, status) {
            $("#conf-diag").modal("hide");
            fnDisplay();
        }

        // build up json
        if ($(this).attr("mode") == "add") {
            ajax.data =  JSON.stringify({
                "type":"mgmt",
                "action":"add",
                "args": {
                    "job": $("#job-tag").val(),
                    "command": $("#tool-tag").val() + " " + $("#args-tag").val(),
                    "disp_id": $("#node-tag").val(),
                    "policy": fnFmtPolicy($("#policy-tag").val())
                }
            });
        } else {
            ajax.data =  JSON.stringify({
                "type":"mgmt",
                "action":"modify",
                "args": {
                    "job": $("#job-tag").val(),
                    "command": $("#tool-tag").val() + " " + $("#args-tag").val(),
                    "policy": fnFmtPolicy($("#policy-tag").val())
                }
            });
        }

        // send add / modify
        $.ajax(ajax);
    });
}

function fnSearch()
{
    confLike = $("#search input").val();
    setC("search", $("#search input").val(), 7);
    confIndex = 0;
    fnDisplay();
}

function fnDirectTo(page) {
    $("#paging li").removeClass("active");
    confIndex = page - 1;
    fnDisplay();
}

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
            confNodes = json.args.result;
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

$(function(){
    fnGetNodes();

    // read cookie
    if (getC("search") != "") {
        $("#search input").val(getC("search"));
        fnSearch();
    } else {
        // display configurations
        fnDisplay();
    }

    $("#add-btn").click(function(){
        var html = "";
        $.each(confNodes, function(k, v){
            html += '<option value="' + v.disp_id + '">' + v.desc + '</option>';
        });
        fnShowDiag({
            mode: "add",
            title: "新增推流",
            tool:  "rtmp_transfer",
            job:   "",
            policy: "is_sleep=1\nres=*;2;*",
            nodes: html
        });
    });

    // press enter key
    $(document).keypress(function(e) {
        if (e.which == 13) {
            fnSearch();
        }
    });
});
