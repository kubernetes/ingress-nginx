 The example to use `nginx.ingress.kubernetes.io/canary-weight` is mentioned at [example](canaryexamples.yaml)
 `nginx.ingress.kubernetes.io/canary-weight: "40"` tells the ingress-nginx to configure Nginx in a way that it proxies 40% of total requests destined to the canary version of app.
 [shell script](counts.sh) can be used to get the output
 Like :
 

<!-- Bash Script to get the output counts is in counts.sh -->
~ ./count.sh #below output shows the requests flow, where 40%(approx) going to cannary pod.
 Total number of requests routed to Pod stable-whoami-6ddd8d7ccc-hk787 are: 57
 Total number of requests routed to Pod canary-whoami-7dbc6bf7dc-tc74n are: 43
~ ./count.sh
 Total number of requests routed to Pod stable-whoami-6ddd8d7ccc-hk787 are: 51
 Total number of requests routed to Pod canary-whoami-7dbc6bf7dc-tc74n are: 49
 
**Note** Canary rules are evaluated in order of precedence. Precedence is as follows: `canary-by-header - > canary-by-cookie - > canary-weight`. Thus, the annotation canary-weight will be ignored if canary-by-header or canary-by-cookie are mentioned.

For notifying the Ingress to route the request to the service specified in the canary Ingress, we need to edit the canary ingress and add following annotation :

`nginx.ingress.kubernetes.io/canary-by-header: "you-can-use-anything-here"`

[example](canaryexamples.yaml):


Then when sending the request to Canary, pass HTTP header `you-can-use-anything-here` set to `always`: [example](canaryexamples.yaml)

If you want your request to be never proxied to the canary backend service then you can set the header to `never`: [example](canaryexamples.yaml)

When the value is absent or set anything other than `never` or `always` then proxying by weight will be used. This works exactly the same for `nginx.ingress.kubernetes.io/canary-by-cookie`.

For example let’s say you want to show the new version of your app to only users under 30 years old. You can then configure

`nginx.ingress.kubernetes.io/canary-by-cookie: "use_under_30_feature"`

and implement a simple change in your backend service that for a given requests checks the age of signed-in user and if the age is under 30 sets cookie `use_under_30_feature` to `always`. This will then make sure all subsequent requests by those users will be proxied by the new version of the app.

[example](canaryexamples.yaml)

References:
 `https://v2-1.docs.kubesphere.io/docs/quick-start/ingress-canary/`
 `https://www.elvinefendi.com/2018/11/25/canary-deployment-with-ingress-nginx.html`
