var Tables = {};

Tables.getNode = function(table) {
  name = table.name.replace(/_/gi, ' ');
  desc = table.description != '' ? table.description : '<span style="color: #999;">no description</span>';
  feeds = table.feeds != 0 ? table.feeds : '<strong>no sources</strong>'
  return `<div class="tcard">
    <div class="tcard-head">
      <span class="tcard-name">` + name + `</span>
      <span class="tcard-type">` + table.type + `</span>
    </div>
    <div class="tcard-desc">` + desc + `</div>
    <div class="tcard-meta">
      <ul>
        <li>members: <strong>` + table.members + `</strong></li>
        <li>news sources: ` + feeds + `</li>
      </ul>
    </div>
    ` + Tables.getMembershipBtn(table.name, table.is_member) + `
    </div>`
}


Tables.getMembershipBtn = function(table, isMember) {
  if (isMember) {
    return `<button type="button" onclick='Tables.toggleMembership("` + table + `",this)' class='black-btn'>member</button>`
  } else {
    return `<button type="button" onclick='Tables.toggleMembership("` + table + `",this)' class='blue-btn'>join</button>`
  }
}

Tables.toggleMembership = function(table, b) {
  btn = $(b)
  btn.prop("disabled", true);
  $.post("/memberships/toggle", {
    table: table
  }).done(function(data, status) {
    if (btn.hasClass("black-btn")) {
      btn.addClass("blue-btn")
      btn.removeClass("black-btn")
      btn.text("join")
    } else {
      btn.addClass("black-btn")
      btn.removeClass("blue-btn")
      btn.text("member")
    }
  }).fail(function(error) {
    Errors.handle(error)
  }).always(function() {
    btn.prop("disabled", false);
  });
}