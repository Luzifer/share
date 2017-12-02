fileURL = undefined

$ ->
  $(window).bind 'hashchange', hashLoad
  hashLoad()
    
hashLoad = ->
  file = window.location.hash.substring(1)
  embedFileInfo(file)

embedFileInfo = (file) ->
  fileURL = file
  $.ajax file,
    method: 'HEAD'
    success: handleEmbed

handleEmbed = (data, status, xhr) ->
  type = xhr.getResponseHeader 'Content-Type'

  console.log fileURL

  $('.show-loading').hide()
  $('.filelink-href').attr 'href', fileURL
  $('.filelink-src').attr 'src', fileURL
  $('.filename').text fileURL.substring(fileURL.lastIndexOf('/') + 1)

  if type.match /^image\//
    $('.show-image').show()
    return

  if type.match /^video\//
    $('.show-video').show()
    return

  $('.show-generic').show()
