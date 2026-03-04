/* Publish by EComposer at 2024-07-28 21:39:25*/
                (function(){
                    const Func = (function() {
                        'use strict';
window.__ectimmers = window.__ectimmers ||{};window.__ectimmers["ecom-omoi68w3xqm"]=  window.__ectimmers["ecom-omoi68w3xqm"] || {};
if(this.settings.link==="lightbox"&&this.settings.lightbox==="yes"&&window.EComModal&&this.$el){var e=this.$el.querySelector("[ecom-modal]");new window.EComModal(e,{cssClass:["ecom-container-lightbox-"+this.id]})}let i=this.$el;if(!i)return;function t(s){const o=s.getBoundingClientRect();return o.top>=0&&o.left>=0&&o.bottom-s.offsetHeight/2<=(window.innerHeight||document.documentElement.clientHeight)&&o.right<=(window.innerWidth||document.documentElement.clientWidth)}function n(){let s=i.querySelector(".ecom-element.ecom-base-image"),o=i.closest(".core__row--columns");s&&(t(s)?(s.classList.add("image-highlight"),o.setAttribute("style","z-index: unset")):(s.classList.remove("image-highlight"),o.setAttribute("style","z-index: 1")))}this.settings.highligh_on_viewport&&window.addEventListener("scroll",function(){n()})

                    });
                    
                        document.querySelectorAll('.ecom-omoi68w3xqm').forEach(function(el){
                            Func.call({$el: el, id: 'ecom-omoi68w3xqm', settings: {"link":"custom","lightbox":"no"},isLive: true});
                        });
                    
                        document.querySelectorAll('.ecom-d0f8rc8ph0i').forEach(function(el){
                            Func.call({$el: el, id: 'ecom-d0f8rc8ph0i', settings: {"link":"none","lightbox":"no"},isLive: true});
                        });
                    

                })();
            
                (function(){
                    const Func = (function() {
                        'use strict';
window.__ectimmers = window.__ectimmers ||{};window.__ectimmers["ecom-uelqmswik69"]=  window.__ectimmers["ecom-uelqmswik69"] || {};
if(!this.$el)return;const e=this.$el,i=this,c=this.settings.layout;let l=e.closest(".core__row--columns");const n=e.querySelector(".ecom-shopify__menu-list--mobile"),s=e.querySelector(".ecom-menu__icon-humber"),m=e.querySelector(".ecom-menu-collapse-close--mobile");let u=e.closest("div.ecom-core.core__block")||"",d=e.closest("div.ecom-column.ecom-core")||"",y=e.querySelectorAll(".ecom-shopify__menu-item--link");for(u&&(u.style.overflow="visible"),y&&n&&y.forEach(function(t){t.addEventListener("click",function(){g()})}),s&&(s.addEventListener("click",v),m.addEventListener("click",g));l;)l.style.zIndex="100",l=l.parentElement.closest(".core__row--columns");function v(){!n||(n.parentNode.style.display="block",n.classList.add("ecom-show"),u&&(u.style.zIndex="100"),d&&(d.style.zIndex="100"),document.querySelector("body").classList.add("ecom-menu-opened"),setTimeout(function(){document.addEventListener("click",p),document.addEventListener("keydown",f)},500))}function p(t){let a=t.target;do{if(a==n)return;a=a.parentNode}while(a);a!=n&&(g(),document.removeEventListener("click",p),document.removeEventListener("keydown",f))}function f(t){(t.isComposing||t.keyCode===27)&&(g(),document.removeEventListener("keydown",f),document.removeEventListener("click",p))}function g(){n.parentNode.style.display="none",n.classList.remove("ecom-show"),u&&(u.style.zIndex="1"),d&&(d.style.zIndex="1"),document.querySelector("body").classList.remove("ecom-menu-opened"),document.removeEventListener("keydown",f),document.removeEventListener("click",p)}let z=e.querySelector('.ecom-shopify__menu-list[data-menu-layout="horizontal"]'),k=null;z&&(k=z.querySelectorAll(".ecom-shopify__menu-item--has-children>.ecom-menu_item>.ecom-element--menu_title")),k&&k.forEach(function(t){t.addEventListener("click",function(a){a.preventDefault()})});function S(){var t=e.querySelectorAll(".ecom-shopify__menu-list .ecom-shopify__menu-item--has-children > .ecom-menu_item, .ecom-shopify__menu-list .ecom-shopify__menu-child-link-item--has-children > .ecom-menu_item"),a=e.querySelectorAll(".ecom-shopify__menu-list--mobile .ecom-shopify__menu-item--has-children > .ecom-menu_item, .ecom-shopify__menu-list--mobile .ecom-shopify__menu-child-link-item--has-children > .ecom-menu_item");if(!!t){var r,b="false",w=e.querySelector(".ecom-shopify__menu-wrapper");if(w&&w.dataset.showAll)var b=w.dataset.showAll;for(r=0;r<t.length;r++){let x=function(o){let _=o.nextElementSibling,h=null;if(o.classList.contains("ecom-item-active")){if(o.classList.remove("ecom-item-active"),_){_.style.maxHeight=null;var $=_.querySelectorAll(".ecom-menu_item");$&&$.forEach(L=>{var E=L.nextElementSibling;E&&(E.style.maxHeight=null),L.classList.remove("ecom-item-active")}),h=o.closest(".ecom-shopify__menu-sub-menu"),h&&(h.style.maxHeight=parseInt(h.style.maxHeight)-_.scrollHeight+"px")}}else o.classList.add("ecom-item-active"),_&&(h=o.closest(".ecom-shopify__menu-sub-menu"),h&&(h.style.maxHeight=parseInt(h.style.maxHeight)+_.scrollHeight+"px"),_.style.maxHeight=_.scrollHeight+"px")};c==="horizontal"&&!i.isLive?t[r].addEventListener("click",function(o){o.preventDefault()}):c==="horizontal"&&i.isLive?t[r].addEventListener("click",function(o){o.stopPropagation()}):(c==="vertical"||!i.isLive)&&(b&&b=="true"&&x(t[r]),t[r].addEventListener("click",function(o){o.preventDefault(),x(this)})),a[r]&&a[r].addEventListener("click",function(o){o.preventDefault(),x(this)})}}}S()

                    });
                    
                        document.querySelectorAll('.ecom-uelqmswik69').forEach(function(el){
                            Func.call({$el: el, id: 'ecom-uelqmswik69', settings: {"layout":"horizontal"},isLive: true});
                        });
                    

                })();
            
;try{
 
} catch(error){console.error(error);}