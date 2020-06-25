<table>
  <tbody>
    <tr>
      <td><img src="https://upload.wikimedia.org/wikipedia/commons/7/75/Dialog-warning-yellow.svg" /></td>
      <td>
        <b>Only the nginx image is meant to be published</b><br/>
        Other images are used as examples for a feature of the ingress controller or to run e2e tests
      </td>
    </tr>
  </tbody>
</table>


Directory | Purpose
------------ | -------------
custom-error-pages | Example of Custom error pages for the NGINX Ingress controller
e2e | Image to run e2e tests
fastcgi-helloserver | FastCGI application for e2e tests
grpc-fortune-teller | grpc server application for the nginx-ingress grpc example
httpbin | A simple HTTP Request & Response Service for e2e tests
nginx | NGINX base image using [alpine linux](https://www.alpinelinux.org)
cfssl | Image to run cfssl commands
