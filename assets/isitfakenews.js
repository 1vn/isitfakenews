$("#url").on("keyup", function(){
    var url = $("#url").val()

    if(!url || url.length == 0){ 
        return
    }

    $("#result").attr("class", "alert alert-warning")
    $("#result").text("Thinking...")
    $.get("/infer", {"url": url}).done(function(data){
        var text;
        var cl;
        
        console.log(data)
        console.log(data.result)
        if(data.result){
            text = "It's fake."
            cl = "alert alert-danger"
        } else {
            text = "It's real."
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