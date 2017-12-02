fileURL = undefined

MSG_NOT_FOUND = 'File not found'
MSG_NOT_PERMITTED = 'Not allowed to access file'
MSG_GENERIC_ERR = 'Something went wrong'

$ ->
  $(window).bind 'hashchange', hashLoad
  hashLoad()
    
hashLoad = ->
  file = window.location.hash.substring(1)
  embedFileInfo(file)

embedFileInfo = (file) ->
  if file == ''
    return handleErrorMessage MSG_NOT_FOUND

  fileURL = file
  $.ajax file,
    method: 'HEAD'
    success: handleEmbed
    error: handleError

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

handleError = (xhr, status) ->
  message = switch xhr.status
    when 404 then MSG_NOT_FOUND
    when 403 then MSG_NOT_PERMITTED
    else MSG_GENERIC_ERR

  handleErrorMessage message

handleErrorMessage = (message) ->
  $('.error').text message
  $('.show-loading').hide()
  $('.show-error').show()
