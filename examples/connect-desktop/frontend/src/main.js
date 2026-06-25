import './style.css';

import { GetEndpointIDString } from '../wailsjs/go/main/App';
import { EventsOn } from "../wailsjs/runtime/runtime";

document.querySelector('#app').innerHTML = `
      <div class="result" id="result">loading endpoint id...</div>
      <div class="messages" id="messages"></div>
`;

let resultElement = document.getElementById("result");

EventsOn("data", ({ data }) => {
  const m = document.getElementById("messages")
  // append to previous output
  m.innerText = m.innerText + data + "\n"
})


function init () {
  // checky little 1s wait just to let the iroh bits properly initialize
  setTimeout(() => {
    try {
      GetEndpointIDString()
        .then((result) => {
          resultElement.innerText = `endpoint id is ${result}`;
        })
        .catch((err) => {
          console.error(err);
        });
    } catch (err) {
      console.error(err);
    }
  }, 1000)
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", init)
} else {
  init()
}
