const http = require('http')
const express = require('express')

const app = express()
const server = http.createServer(app)

let startTime = null
app.get('/', (req, resp) => {
  server.getConnections((err, count) => {
    if (err) return

    const time = new Date().getTime()
    if (startTime == null) startTime = time

    console.log(time - startTime, count)
  })

  resp.status(200).send()
})

const port = process.env.PORT
server.listen(port, () => {
  console.error(`listening on port ${port}`)
})