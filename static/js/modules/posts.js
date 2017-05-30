// this lib deals with posts and posting forms
var Posts = {};
// posts.sorting = 'top';
Posts.identity = 'real';
Posts.getNode = function(post) {
  // TODO: likes, comments, links, imgae
  created = new Date(post.created);
  interval = ago.format(created, 'en');
  likeClass = post.liked ? "liked" : "";
  _likes = post.likes == 0 ? "" : post.likes
  _comments = post.comments == 0 ? "" : post.comments
  imgsrc = 'https://s3.amazonaws.com/spacecafe-users/' + post.user_id + '_44.png';
  if (post.user_id == 0) {
    imgsrc = 'https://s3.amazonaws.com/spacecafe-users/' + post.nimg + '.png'
  }
  return '<div class="post post-' + post.id + ' ' + likeClass + '"><img class="userimg" src="' + imgsrc + '""><div class="head"><div class="_head-info"><span class="username">' + post.user_name + '</span><span class="date"><a href="/post/' + post.id + '">' + interval + '</a></span></div><div class="_head-line"></div></div><div class="body">' + post.body + '</div><div class="ctrl"><div class="_content"><span class="action like-btn" onclick="togglePostLike(' + post.id + ')">like</span><span class="count _likes' + _likes + '">' + _likes + '</span><span class="action" onclick="showPost(' + post.id + ')">comment</span><span class="count _comments' + _comments + '">' + _comments + '</span></div><div class="_line"></div></div></div>'
}
// <span class="action"><a href="/t/' + post.table_name + '">' + post.table_name + '</a></span>


Posts.growTextArea = function(id) {
  area = $('#' + id).get(0)
  if (area.attributes["rows"] > 1) {
    return;
  }
  area.setAttribute("rows", 2);
  autosize(area);
  area.classList.add("light-placeholder");
  ctrl = $("#" + id + "-ctrl").get(0)
  ctrl.classList.remove("hide");
}

Posts.shrinkTextArea = function(id) {
  area = $('#' + id).get(0)
  area.value = "";
  autosize.destroy(area);
  area.setAttribute("rows", 1);
  area.classList.remove("light-placeholder");
  ctrl = $("#" + id + "-ctrl").get(0);
  ctrl.classList.add("hide");
}

Posts.beAnonymous = function() {
  Posts.identity = 'anonymous';
  // setPostingHint("anonymous")
  $(".identity").addClass("hide")
  $(".anonymous-identity").removeClass("hide")
  $(".identity-flag").val("n")
}


Posts.beReal = function() {
  Posts.identity = 'real';
  // setPostingHint("real")
  $(".identity").addClass("hide")
  $(".real-identity").removeClass("hide")
  $(".identity-flag").val("r")
}