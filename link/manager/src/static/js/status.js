var pslist = [];
var psIndex = 0;
var psRows = 20;
var psTotal = 0;
var psLike = "";
var psTimer;
var psNodes = [];

var traceSize = 5000;

function fnDirectTo(page) {
    $("#paging li").removeClass("active");
    psIndex = page - 1;
    fnPS();
}

function fnRef() {
    var html = "";
    $.each(pslist, function(k, v){
        html += "<tr>";
        html += $("#tpl-row").tmpl({
            job: v.job,
            start_time: v.start_time,
            is_sleep: v.is_sleep,
            is_modified: v.is_modified,
            status: v.status,
            nodes: fnFmtNodes(v.job, v.disp_id)
        }).html();
        html += "</tr>";
    });
    $("#job-list").html(html);
    $(".sleep-switch").bootstrapSwitch();
    $(".sleep-switch").bootstrapSwitch("onSwitchChange", function(e, state){
        // ajax json
        var ajax = {
            type: 'POST',
            contentType: "application/json; charset=utf-8",
            url: '/api/action'
        };

        // build wake/sleep json strings
        if (state == true) {
            // wake
            ajax.data =  JSON.stringify({
                "type":"mgmt",
                "action":"wake",
                "args": {"job": $(this).attr("job")}
            });
            ajax.error = function (xhr, desc, err) {
                alert(xhr.responseText);
            }
        } else {
            // sleep
            ajax.data =  JSON.stringify({
                "type":"mgmt",
                "action":"sleep",
                "args": {"job": $(this).attr("job")}
            });
            ajax.error = function (xhr, desc, err) {
                alert(xhr.responseText);
            }
        }

        // send sleep / wake request
        $.ajax(ajax);

    });

    // load pagination
    html = ""
    for (i = 0; i < Math.ceil(psTotal / psRows); i++) {
        var active = "0";
        if (i == psIndex) {
            active = "1";
        }
        html += $("#tpl-paging").tmpl({
            index : i + 1,
            active: active
        }).html();
    }
    $("#paging").html(html);

    // add empty tips
    if (pslist.length == 0) {
        $("#job-list").html("<tr><td>没有查询到进程信息</td><td></td><td></td><td></td><td></td><td></td></tr>");
    }
} // fnRef

function fnPS() {
    var data = {
        "type":"mgmt",
        "action":"ps",
        "args": {
            "index":psIndex.toString(),
            "rows":psRows.toString()
        }
    };
    if (psLike != "") {
        data.args.like = psLike.toString();
    }

    $.ajax({
        type: 'POST',
        contentType: "application/json; charset=utf-8",
        url: '/api/action',
        data: JSON.stringify(data),
        success: function (json, status) {
            pslist = json.args.result;
            psTotal = json.args.total;
            fnRef();
            clearTimeout(psTimer);
            psTimer = setTimeout(fnPS, 5000);
        }
    });
} // fnPS

function fnSearch()
{
    psLike = $("#search input").val();
    setC("search", $("#search input").val(), 7);
    psIndex = 0;
    fnPS();
}

function fnTrace(job)
{
    var traceFunc = function(successCB, errorCB) {
        $.ajax({
            type: 'POST',
            contentType: "application/json; charset=utf-8",
            url: '/api/action',
            data: JSON.stringify({
                "type":"mgmt",
                "action":"tail",
                "args": {
                    "job": job,
                    "size": traceSize.toString()
                }
            }),
            success: successCB,
            error: errorCB
        });
    }

    traceFunc(
        function (json, status) {
            var obj = {
                title: job + " - 输出信息",
                trace: json.args.result
            }
            $("#trace-diag").modal("show");
            var html = $("#tpl-trace").tmpl(obj).html();
            $("#trace-diag .conf-content").html(html);
            $("#trace-diag").on("shown.bs.modal", function(e){
                $('#trace-diag textarea').scrollTop($('#trace-diag textarea')[0].scrollHeight);
            });
            pollTimer = setInterval(traceFunc, 3000,
                                   function(msg, status){
                                       $("#trace-diag textarea").text(msg.args.result);
                                   }, null);
            $("#trace-diag").on("hide.bs.modal", function(e){
                clearInterval(pollTimer);
            });
        },
        function (xhr, desc, err) {
            alert(xhr.responseText);
        }
    );
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
            psNodes = json.args.result;
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
} // fnGetNodes

function fnFmtNodes(job, id)
{
    var html = "";
    $.each(psNodes, function(k, v){
        html += '<option ';
        html += (id == v.disp_id ? "selected" : "");
        html += ' job="' + job + '" disp="' + v.disp_id + '">';
        html += v.desc + '</option>';
    });
    return html;
} // fnFmtNodes

function fnTrespass(dom)
{
    // stop timeout
    clearTimeout(psTimer);

    var job = $(dom).children("option:selected").attr("job");
    var disp = $(dom).children("option:selected").attr("disp");

    var data = {
        "type":"mgmt",
        "action":"trespass",
        "args":{"job":job, "disp_id":disp}
    };

    // send req
    $.ajax({
        type: 'POST',
        contentType: "application/json; charset=utf-8",
        url: '/api/action',
        data: JSON.stringify(data),
        success: function (json, status) {
            // restore timeout
            psTimer = setTimeout(fnPS, 5000);
        },
        error: function (xhr, desc, err) {
            alert(xhr.responseText);

            // restore timeout
            fnPS();
            psTimer = setTimeout(fnPS, 5000);
        }
    });
} // fnTrespass

$(function(){
    fnGetNodes();

    // read cookie
    if (getC("search") != "") {
        $("#search input").val(getC("search"));
        fnSearch();
    } else {
        fnPS();
    }

    // press enter key
    $(document).keypress(function(e) {
        if (e.which == 13) {
            fnSearch();
        }
    });
})
