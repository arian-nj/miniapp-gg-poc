import './style.css'
import { setupCounter } from './counter.ts'
import WebApp from '@twa-dev/sdk'

WebApp.ready()

fetch("/api/open?" + WebApp.initData)

let full_name = ""
let username = WebApp.initDataUnsafe.user?.username == undefined ? "no username" : `@${WebApp.initDataUnsafe.user.username}`

if (WebApp.initDataUnsafe.user?.first_name !== undefined) {
	full_name += WebApp.initDataUnsafe.user?.first_name
}


if (WebApp.initDataUnsafe.user?.last_name !== undefined) {
	full_name += WebApp.initDataUnsafe.user?.last_name
}

document.querySelector<HTMLDivElement>('#app')!.innerHTML = `
  <div>
	  <h1>Hello ${full_name}</h1>
	  <h3>${username}</h3>
	  <p>Fuck You</p>
  </div>
`

setupCounter(document.querySelector<HTMLButtonElement>('#counter')!)
