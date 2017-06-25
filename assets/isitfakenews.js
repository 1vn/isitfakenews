console.log("yolo")
$("#url").on("keyup", function(){
    var url = $("#url").val()
    $.get("http://localhost:5555/infer", {"url": url}).done(function(data){
        var text;
        var cl;
        if(data.result){
            text = "YES"
            cl = "alert alert-danger"
        } else {
            text = "NO, THIS IS PROBABLY REAL"
            cl = "alert alert-success"
        }
        $("#result").addClass(cl)
        $("#result").text(text)
        $("#correct").css({"display": "block"})
    })
})