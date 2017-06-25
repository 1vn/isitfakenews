console.log("yolo")
$("#url").on("keyup", function(){
    var url = $("#url").val()
    $("#result").attr("class", "alert alert-warning")
    $("#result").text("Thinking...")
    $.get("/infer", {"url": url}).done(function(data){
        var text;
        var cl;
        if(!data.result){
            text = "YES"
            cl = "alert alert-danger"
        } else {
            text = "NO"
            cl = "alert alert-success"
        }
        $("#result").attr("class", cl)
        $("#result").text(text)
        $("#correct").css({"display": "block"})
    })
})

$("#yesButton").on("click", function(){
    var url = $("#url").val()
    $("#correct").css({"display": "none"})
    $.post("/correct", {"correct": true, "url": url})
})


$("#noButton").on("click", function(){
    var url = $("#url").val()
    $("#correct").css({"display": "none"})
    $.post("/correct", {"correct": false, "url": url})
})