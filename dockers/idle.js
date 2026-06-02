function f() { setTimeout(f, 60*60*1000) }
function exit() {
    console.log('exiting')
    process.exit()
}
process.on('SIGTERM', exit)
process.on('SIGINT', exit)
f()
