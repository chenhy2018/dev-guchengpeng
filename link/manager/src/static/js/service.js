var svclist = [];

function fnRef()
{
    $("#grid").html("");
    $.each(svclist, function(k, v){

        var title;
        var version;
        var ip;
        var port = "";
        var detail;

        $.each(v, function(kk, vv){
            switch (kk) {
            case "module":
                switch (vv) {
                case "mgmt":
                    title = "管理服务";break;
                case "dispatcher":
                    title = "调度节点";break;
                case "worker":
                    title = "运行节点";break;
                default:
                    title = "未知";
                }
                break;
            case "version":
                version = vv; break;
            case "ip":
                ip = vv; break;
            case "port":
                port = vv; break;
            default:
            }
        });

        var html = $("#tpl-blk").tmpl({
            title: title,
            content: "v" + version + "<br>" + ip + (port == "" ? "" : ":" + port),
            fdetail: k,
            fupgrade: null
        }).html();

        $(".grid").append(html);
    });
}

function fnVersion()
{
    $.ajax({
        type: 'POST',
        contentType: "application/json; charset=utf-8",
        url: '/api/action',
        data: JSON.stringify({
            "type":"mgmt",
            "action":"version",
            "args":{}
        }),
        success: function (json, status) {
            svclist = json.args.result;
            fnRef();
        },
        error: function (xhr, desc, err) {
            switch (xhr.status){
            case 403:
                alert("无授权查看该页面");
                window.location.href='/';
                break;
            default:
                alert("不能获取服务节点信息");
            }
        }
    });
}

function fnDetail(index)
{
    var detail = "";
    $.each(svclist[index], function(k, v){
        switch (k) {
        case "module":
        case "version":
        case "ip":
        case "port":
            break;
        case "worker_id":
            detail += "\n服务ID:\n" + v;
            break;
        case "extra":
            detail += "\nPub Tools版本:\n" + v;
            break;
        default:
            detail += "\n" + v;
        }
    });
    if (detail == "") {
        detail += "无详细信息\n";
    }

    var obj = {
        title: "详细信息",
        info: detail
    };
    $("#info-diag").modal("show");
    var html = $("#tpl-info").tmpl(obj).html();
    $("#info-diag .conf-content").html(html);
}

$(function(){
    fnVersion();
});
