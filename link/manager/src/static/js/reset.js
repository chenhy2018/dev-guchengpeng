$(function(){

    $("#reset-btn").click(function(){
        // check new pwd
        if ($("#name").val() == "" || $("#pass").val() == "" || $("#cpass").val() == "" ||  $("#opass").val() == "") {
            alert("表单内容不可为空");
            return;
        }
        if ($("#pass").val() != $("#cpass").val()) {
            alert("新密码与密码确认不一致");
            return;
        }
        if ($("#pass").val() == $("#opass").val()) {
            alert("新密码与原密码相同");
            return;
        }

        // send POST
        $.ajax({
            type: 'POST',
            contentType: "application/json; charset=utf-8",
            url: '/api/action',
            data: JSON.stringify({
                "type":"mgmt",
                "action":"reset",
                "args": {
                    "uid": $("#name").val(),
                    "opassword": $("#opass").val(),
                    "password": $("#cpass").val()
                }
            }),
            success: function (msg, status) {
                alert("密码更改成功！");
                window.location.href='/';
            },
            error: function (xhr, desc, err) {
                alert("更新失败，可能用户名和原始密码不匹配");
            }
        });
    });
});
