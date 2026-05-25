# SitePackage diff

## Site `[100]` Scripted

- ➖ `Sites[SiteId=100].SiteSettings.Scripts[1]` = (object, 3 keys)
- ➕ `Sites[SiteId=100].SiteSettings.Scripts[1]` = (object, 3 keys)

**`Sites[SiteId=100].SiteSettings.Scripts[Title=amount-helper].Body`**
```diff
--- old
+++ new
@@ -1,5 +1,5 @@
-// auto-fills amount
-function calc(price, qty) {
-  return price * qty;
+// auto-fills amount, with discount
+function calc(price, qty, discount) {
+  return price * qty * (1 - discount);
 }
 
```

