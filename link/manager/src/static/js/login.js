$(function(){

    // auth check
    var fnAuth = function(){
        $.ajax({
            type: 'POST',
            contentType: "application/json; charset=utf-8",
            url: '/api/action',
            data: JSON.stringify({
                "type":"mgmt",
                "action":"auth",
                "args": {}
            }),
            success: function (msg, status) {
                window.location.href='/view/service';
            },
            error: function (xhr, desc, err) {
                //alert(xhr.responseText);
            }
        });
    }();

    // login button
    $("#login-btn").click(function(){
        $.ajax({
            type: 'POST',
            contentType: "application/json; charset=utf-8",
            url: '/api/action',
            data: JSON.stringify({
                "type":"mgmt",
                "action":"login",
                "args": {
                    "uid": $("#login-name").val(),
                    "password": $("#login-pass").val()
                }
            }),
            success: function (msg, status) {
                window.location.href='/view/service';
            },
            error: function (xhr, desc, err) {
                alert("用户名或密码错误");
            }
        });
    });

    // press enter key
    $(document).keypress(function(e) {
        if (e.which == 13) {
            $("#login-btn").click();
        }
    });

});
