var comments = {};
comments.objs = [];
comments.tree = '';
comments.lastBuilt = 0;
comments.idIdxMap = {};
comments.idxDepthMap = {};
comments.mode = cookies.get("comments_mode") == '' ? 'tree' : cookies.get("comments_mode");
comments.postId;
comments.threadId = '';
// comments.replyToId = 0;
// comments.replyToDepth = 0;
comments.done = false;

comments.clear = function() {
  comments.objs = [];
  comments.tree = '';
}

comments.clearMaps = function() {
  comments.idIdxMap = {};
  comments.idxDepthMap = {};
}

comments.changeMode = function(mode) {
  if (comments.mode == mode) {
    return
  }
  comments.mode = mode;
  cookies.set('comments_mode', mode, Infinity);
  $("#comments-wrapper").empty();
  comments.fetch()
}


comments.fetch = function() {
  comments.done = false;
  $("#comments-wrapper").removeClass("comments-mode-tree")
  $("#comments-wrapper").addClass("comments-mode-" + comments.mode)
  if (comments.mode == 'tree') {
    $.post("/comments", {
      post_id: comments.postId,
      thread_id: comments.threadId
    })
      .done(function(data) {
        // console.log(data)
        // $("#comments").append(getCommentNode(data))
        // lastPostCount = data.posts.length;
        // comments = '';
        // for (var i = 0; i < data.length; i++) {
        //  // if (data[i].depth == 0) {}
        //  comments = comments + getCommentNode(data[i]) 
        //  // comments = comments + '<div class="comment-childern">'
        // }
        if (data.length == 0) {
          return
        }
        comments.objs = data;
        comments.build();
        $("#comments-wrapper").append(comments.tree)
        comments.clear();
      // if (!hasScrollbar() && lastPostCount == 4) {loadPosts()}
      })
      .fail(function(data) {
        // console.log(data)
      });
  } else {
    $.post("/comments/flat", {
      post_id: comments.postId,
      thread_id: comments.threadId
    })
      .done(function(data) {
        // console.log(data)
        // $("#comments").append(getCommentNode(data))
        // lastPostCount = data.posts.length;
        // comments = '';
        if (data.length == 0) {
          return
        }
        for (var i = 0; i < data.length; i++) {
          data[i].depth = 1;
        //  // if (data[i].depth == 0) {}
        //  comments = comments + getCommentNode(data[i]) 
        //  // comments = comments + '<div class="comment-childern">'
        }
        comments.objs = data;
        comments.build();
        $("#comments-wrapper").append(comments.tree)
        comments.clear();
      // if (!hasScrollbar() && lastPostCount == 4) {loadPosts()}
      })
      .fail(function(data) {
        // console.log(data)
      });
  }


}

comments.moveUnderParent = function(btn, parentId, id) {
  var idx = comments.idIdxMap[id]
  var parentIdx = comments.idIdxMap[parentId];
  var middleIdxs = [];

  // get comments on before the current and have the same depth
  for (var i = parentIdx + 1; i < idx; i++) {
    if (comments.idxDepthMap[idx] == comments.idxDepthMap[i]) {
      middleIdxs.push(i);
    }
  }

  if ($("#comment-wrapper-" + id)[0].classList.contains("comment-moved-under-parent")) {
    $.map(middleIdxs, function(idx, i) {
      $(".comment-wrapper-idx-" + idx).removeClass("hide")
    })
    $("#comment-wrapper-" + id).removeClass("comment-moved-under-parent")
    btn.innerHTML = '[<strong>&#8673;</strong>]'
  } else {
    $.map(middleIdxs, function(idx, i) {
      $(".comment-wrapper-idx-" + idx).addClass("hide")
    })
    $("#comment-wrapper-" + id).addClass("comment-moved-under-parent")
    btn.innerHTML = '[<strong>&#8675;</strong>]'
  }
}

comments.build = function(idx) {
  if (comments.done) {
    return;
  }
  idx = idx || 0
  comments.idIdxMap[comments.objs[idx].id] = idx;
  comments.idxDepthMap[idx] = comments.objs[idx].depth;

  comments.tree = comments.tree + '<div class="comment-wrapper comment-wrapper-depth-' + comments.objs[idx].depth + ' comment-wrapper-idx-' + idx + '" id="comment-wrapper-' + comments.objs[idx].id + '">'
  comments.tree = comments.tree + comments.getCommentNode(comments.objs[idx])

  comments.tree = comments.tree + '<div class="comment-childern" id="comment-childern-' + comments.objs[idx].id + '">'

  if (idx == comments.objs.length - 1) { // end of tree
    comments.tree = comments.tree + '</div></div>'
    comments.done = true;
    return
  }

  if (comments.objs[idx + 1].depth == comments.objs[idx].depth) { // same level
    // close previous box and start from the next
    comments.tree = comments.tree + '</div></div>'
    comments.build(idx + 1)
  }
  if (comments.objs[idx + 1].depth > comments.objs[idx].depth) { // child
    // stay in box and start from the next
    comments.build(idx + 1)

    // then close the box
    comments.tree = comments.tree + '</div></div>'
    comments.build(comments.lastBuilt + 1)
  }
  if (comments.objs[idx + 1].depth < comments.objs[idx].depth) { // higher level
    // close current box
    comments.tree = comments.tree + '</div></div>'
    comments.lastBuilt = idx;
    return // return to parent box to close itself then start from the next
  }
}


comments.toggleLike = function(id) {
  if ($("#comment-" + id)[0].classList.contains("liked")) {
    $.post("/comments/unlike", {
      comment_id: id
    })
    $("#comment-" + id).removeClass("liked")
    return;
  }
  $.post("/comments/like", {
    comment_id: id
  })
  $("#comment-" + id).addClass("liked")
}


comments.showReplies = function(btn, id, depth) {
  if ($("#comment-wrapper-" + id)[0].classList.contains("replies-visible")) {
    $("#comment-wrapper-" + id).removeClass("replies-visible")
    $("#comment-childern-" + id).addClass("hide")
    btn.innerHTML = 'show replies'
    return
  } else if ($("#comment-childern-" + id)[0].innerHTML != '') {
    $("#comment-childern-" + id).removeClass("hide")
    btn.innerHTML = 'hide replies';
    $("#comment-wrapper-" + id).addClass("replies-visible")
    return
  }



  $.post("/comments/replies", {
    parent_id: id
  })
    .done(function(data) {
      // console.log(data)
      // $("#comments").append(getCommentNode(data))
      // lastPostCount = data.posts.length;
      // comments = '';
      for (var i = 0; i < data.length; i++) {
        data[i].depth = depth + 1;
      //  // if (data[i].depth == 0) {}
      //  comments = comments + getCommentNode(data[i]) 
      //  // comments = comments + '<div class="comment-childern">'
      }
      comments.objs = data;
      comments.build();
      $("#comment-childern-" + id).append(comments.tree)
      $("#comment-childern-" + id).removeClass("hide")
      comments.clear();
      btn.innerHTML = 'hide replies';
      $("#comment-wrapper-" + id).addClass("replies-visible")
    // if (!hasScrollbar() && lastPostCount == 4) {loadPosts()}
    })
    .fail(function(data) {
      // console.log(data)
    });
}

comments.getCommentNode = function(comment) {
  // console.log(comment)
  // TODO: likes, comments, links, imgae
  // TODO: posts older than a day - month day
  // TODO: add tooltip to date
  created = new Date(comment.created)
  interval = ago.format(created, 'en')

  moveUnderParentBtn = ''
  if ((comments.idIdxMap[comment.id] - comments.idIdxMap[comment.parent_id]) > 1) {
    moveUnderParentBtn = '<span class="act" onclick="comments.moveUnderParent(this,' + comment.parent_id + ',' + comment.id + ')">[<strong>&#8673;</strong>]</span>'
  }

  showReplyBtn = ''
  if (comment.replies > 0) {
    showReplyBtn = '<span class="action action-show-replies" onclick="comments.showReplies(this,' + comment.id + ',' + comment.depth + ')">show replies ' + comment.replies + '</span>'
  }

  return '<div class="post comment-depth-' + comment.depth + '" id="comment-' + comment.id + '"><div class="head"><span class="act" onclick="comments.toggle(this,' + comment.id + ')">[–]</span>' + moveUnderParentBtn + '<span class="username">' + comment.user_name + '</span><a class="date" href="/thread/' + comment.id + '">' + interval + '</a></div><div class="body">' + comment.body + '</div><div class="ctrl"><span class="action" onclick="comments.toggleLike(' + comment.id + ')">Like</span><span class="count">30</span><span class="action" onclick="comments.replyTo(' + comment.id + ', ' + comment.depth + ')">Reply</span>' + showReplyBtn + '<span class="action">' + comment.table_name + '</span></div><div id="comment-reply-area-' + comment.id + '"></div></div>'
}


comments.getReplyForm = function(id, depth) {
  return `<div class="post-form" id="reply-form">
        <div class="real-identity identity">
          <img class="usr" src="/static/img/15.jpg">
          <img onclick="postAnonymous()" class="usr-small" src="/static/img/2.jpg">
        </div>

        <div class="anonymous-identity identity hide">
          <img class="usr" src="/static/img/2.jpg">
          <img onclick="postReal()" class="usr-small" src="/static/img/15.jpg">
        </div>

        <div class="anonymous-only identity hide">
          <img class="usr" src="/static/img/15.jpg">
          <img style="cursor: default;opacity: 0.5" class="usr-small" src="/static/img/2.jpg">
        </div>

        <form>
          <textarea id="reply-area" onclick="Posts.growTextArea('reply-area')" rows="1" name="body" placeholder="Write a reply..."></textarea>
          <div class="hide" id="reply-area-ctrl">
            <span class="hint posting-identity-hint">Posting as {{.User.Name}}</span>
            <button class="blue-btn" type="button" onclick="reply(` + id + `,` + depth + `)">Post</button>
          </div>
        </form>
      </div>`;
}

comments.replyTo = function(id, depth) {
  event.stopPropagation();
  $("#comment-reply-area-" + id).append(comments.getReplyForm(id, depth))
// $("#reply-form")[0].classList.remove("hide")
}


// TODO
// get full comment node
// manage idex
// TODO: re-show count after show and hide
// TODO: manage fetching large number of comments or replies
// TODO: read more in comments
// TODO: manage long threads (continue this thread)
comments.new = function() {
  $.post("/comments/new", {
    body: $("#comment-area")[0].value,
    parent_id: "p:" + comments.postId
  })
    .done(function(data) {
      Posts.shrinkTextArea("comment-area")
      data.depth = 0;
      c = '<div class="comment-wrapper comment-wrapper-depth-' + 1 + ' comment-wrapper-idx-' + 0 + '" id="comment-wrapper-' + data.id + '">' + comments.getCommentNode(data) + '</div>'
      $("#comments-wrapper").prepend(c)
    })
    .fail(function(data) {
      // console.log(data)
    });
}

comments.toggle = function(btn, id) {
  if ($("#comment-wrapper-" + id)[0].classList.contains("comment-collapsed")) {
    btn.innerHTML = '[–]'
    $("#comment-wrapper-" + id).removeClass("comment-collapsed")
  } else {
    btn.innerHTML = '[+]'
    $("#comment-wrapper-" + id).addClass("comment-collapsed")
  }
}



// http://stackoverflow.com/questions/17687410/when-will-proper-stack-traces-be-provided-on-window-onerror-function
// window.onerror = function(message, file, line, column, errorObj) {
//   if (errorObj !== undefined) //so it won't blow up in the rest of the browsers
//     console.log('Error: ' + errorObj.stack);
// }