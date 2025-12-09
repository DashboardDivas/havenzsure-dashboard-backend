# RBAC & Shop Enforcement Specification

This document defines role-based access control (RBAC) and shop-level enforcement rules for the HavenzSure Dashboard.

**Status Legend**

- ✅ Implemented in backend (UserService / ShopService / DB)
- ⚠️ Partially implemented (logic exists but differs from spec, e.g., missing shop scoping)
- ❌ Not implemented yet in backend

---

## 1. Roles Overview

| Role           | Description                                 | Requires Shop? (Spec) | Status                                                        |
| -------------- | ------------------------------------------- | --------------------- | ------------------------------------------------------------- |
| **SuperAdmin** | System-level owner with unrestricted access | ❌ No                 | ⚠️ No backend constraint (may or may not have a shop)         |
| **Admin**      | Manager of a single shop                    | ✔ Yes                 | ⚠️ Enforced on create in service; DB column is still nullable |
| **Adjuster**   | Inspection staff belonging to a shop        | ✔ Yes                 | ⚠️ Enforced on create in service; DB column is still nullable |
| **Bodyman**    | Repair staff belonging to a shop            | ✔ Yes                 | ⚠️ Enforced on create in service; DB column is still nullable |

**Global rule (spec):** All non-superadmin users must belong to a shop (`shop_id` cannot be NULL).  
**Backend:** ❌ Not enforced in DB or service; `shop_id` is nullable and not role-scoped.

---

## 2. User Visibility Rules

### 2.1 Who can _see_ which users?

| Target User →            | SuperAdmin (can see?) | Admin (can see?)                      | Status         |
| ------------------------ | --------------------- | ------------------------------------- | -------------- |
| **SuperAdmin**           | ✔ Visible             | ❌ Hidden                             | ✅ Implemented |
| **Admin (any shop)**     | ✔ Visible             | ✔ Visible (read-only if not own shop) | ✅ Implemented |
| **Staff in own shop**    | ✔ Visible             | ✔ Visible                             | ✅ Implemented |
| **Staff in other shops** | ✔ Visible             | ❌ Hidden                             | ✅ Implemented |

**Backend:** ✅ Implemented – admin visibility is enforced via `canViewUser` in both `ListUsers` and `GetUserByID`.

---

## 3. User Action Permissions

### 3.1 Create User

| Action                    | SuperAdmin (spec) | Admin (spec)                      | Status         |
| ------------------------- | ----------------- | --------------------------------- | -------------- |
| Create SuperAdmin         | ✔                 | ❌                                | ✅ Implemented |
| Create Admin              | ✔                 | ❌                                | ✅ Implemented |
| Create Adjuster / Bodyman | ✔                 | ✔ (own shop only)                 | ✅ Implemented |
| Assign `shop_id`          | ✔ Any shop        | ✔ Only own shop                   | ✅ Implemented |
| Assign roles              | ✔ Any role        | ❌ Cannot assign admin/superadmin | ✅ Implemented |

Notes:

- Role-based creation constraints (`admin` cannot create `admin` or `superadmin`) are enforced.
- Shop-based constraints are enforced:
  - Non-superadmin users must have a shop on create.
  - Admins can only create staff in **their own** shop; SuperAdmins may assign any shop or none (for SuperAdmins).

---

### 3.2 Update User

| Target or Field             | SuperAdmin (spec)     | Admin (spec)                 | Status                                                                                                              |
| --------------------------- | --------------------- | ---------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| Update SuperAdmin           | ✔                     | ❌                           | ✅ Implemented – admins cannot manage superadmins                                                                   |
| Update Admin                | ✔                     | ❌                           | ✅ Implemented – admins cannot manage other admins (only themselves)                                                |
| Update own profile          | ✔                     | ✔                            | ✅ Implemented – both can update own fields (with role restrictions)                                                |
| Update staff in own shop    | ✔                     | ✔                            | ✅ Implemented                                                                                                      |
| Update staff in other shops | ✔                     | ❌                           | ✅ Implemented – blocked by shop-based checks in `canManageUser`                                                    |
| Change user role            | ✔ Any role            | ❌ (only adjuster ↔ bodyman) | ✅ Implemented – admins can only switch between non-admin/non-superadmin roles; cannot touch admin/superadmin roles |
| Change user shop            | ✔                     | ❌                           | ✅ Implemented – SuperAdmins may change ShopCode; Admin updates with ShopCode are rejected                          |
| Set `shop_id` to NULL       | ✔ Only for superadmin | ❌                           | ⚠️ Not supported via service; clearing `shop_id` is not allowed for any role (DB column is still nullable)          |

Notes:

- `canManageUser` now enforces both **role-based** and **shop-based** permissions for Admins.
- `checkFieldUpdatePermission` prevents Admins from:
  - changing shop assignment,
  - promoting anyone to `admin` or `superadmin`,
  - changing their own role.
- Clearing `shop_id` is not supported via the service for now; this invariant should be maintained at the DB or migration level if needed.

---

### 3.3 Deactivate / Reactivate User

| Target User        | SuperAdmin (spec) | Admin (spec) | Status                                                                |
| ------------------ | ----------------- | ------------ | --------------------------------------------------------------------- |
| SuperAdmin         | ✔                 | ❌           | ✅ Implemented – admins cannot manage superadmins                     |
| Admin              | ✔                 | ❌           | ✅ Implemented – admins cannot manage other admins                    |
| Staff (own shop)   | ✔                 | ✔            | ✅ Implemented – admins can deactivate/reactivate staff in their shop |
| Staff (other shop) | ✔                 | ❌           | ✅ Implemented – blocked by shop-based checks in `canManageUser`      |
| Self-deactivation  | ❌                | ❌           | ✅ Implemented – service explicitly blocks deactivating yourself      |

Notes:

- `DeactivateUser` / `ReactivateUser` rely on `canManageUser` for both role and shop scoping and explicitly forbid self-deactivation/reactivation.

---

## 4. Shop Management Rules

### 4.1 Shop Visibility

| Shop Information           | SuperAdmin (spec) | Admin (spec)               | Status                                                                      |
| -------------------------- | ----------------- | -------------------------- | --------------------------------------------------------------------------- |
| View all shops             | ✔ Full            | ✔ Full                     | ✅ Implemented – `/shops` lists all shops for any caller; no RBAC filtering |
| View shop details          | ✔ Full            | ✔ Full                     | ✅ Implemented – `/shops/{id}` returns any shop; no RBAC filtering          |
| View shop internal metrics | ✔ Full            | ✔ Only own shop (optional) | ❌ Not implemented – no metrics layer or shop-based restriction yet         |

_(Admins may view details of all shops, since they are non-sensitive contact/business info.)_

Backend note:  
Current `ShopService` and `shop.Handler` **do not receive the actor** and perform **no role/shop RBAC checks**. Any authenticated caller reaching those routes sees the same results.

---

### 4.2 Shop Edit Permissions

| Action                   | SuperAdmin (spec) | Admin (spec) | Status                                                                               |
| ------------------------ | ----------------- | ------------ | ------------------------------------------------------------------------------------ |
| Create shop              | ✔                 | ❌           | ❌ Not implemented – no role-based restriction at service/handler level              |
| Edit any shop            | ✔                 | ❌           | ❌ Not implemented – anyone hitting `PUT /shops/{id}` goes through the same path     |
| Edit own shop            | ✔                 | ✔            | ⚠️ Technically allowed, but not scoped; there is no “own shop only” concept enforced |
| Delete / deactivate shop | ✔                 | ❌           | ❌ Not implemented – no delete/deactivate logic in current shop service/handler      |

---

## 5. WorkOrder Management Rules

> **Important:** Based on the files provided, there is currently **no WorkOrder service/handler code** in the backend. The following table reflects the **intended** design, but is **not implemented yet**.

### 5.1 Visibility

| WorkOrder Scope  | SuperAdmin (spec) | Admin (spec)                       | Adjuster / Bodyman (spec) | Status                                                                 |
| ---------------- | ----------------- | ---------------------------------- | ------------------------- | ---------------------------------------------------------------------- |
| All shops        | ✔                 | ❌ (optional read-only for search) | ❌                        | ❌ Not implemented – WorkOrder service not present in backend code yet |
| Own shop         | ✔                 | ✔                                  | ✔                         | ❌ Not implemented                                                     |
| Assigned to self | ✔                 | ✔                                  | ✔ (“My Work”)             | ❌ Not implemented                                                     |

### 5.2 Actions

| Action                | SuperAdmin (spec) | Admin (spec)    | Adjuster (spec)  | Bodyman (spec)   | Status             |
| --------------------- | ----------------- | --------------- | ---------------- | ---------------- | ------------------ |
| Create workorder      | ✔ Any shop        | ✔ Own shop      | ✔ Own shop       | ❌               | ❌ Not implemented |
| Update workorder      | ✔ Any shop        | ✔ Own shop      | ✔ Limited fields | ✔ Limited fields | ❌ Not implemented |
| Assign workorder      | ✔                 | ✔ Own shop only | ✔ Own shop only  | ❌               | ❌ Not implemented |
| Change workorder shop | ✔                 | ✔ Own shop only | ✔ Own shop only  | ❌               | ❌ Not implemented |

---

## 6. Backend Enforcement Summary

| Layer                | Enforcement (spec)                                                     | Backend Status                                                                                              |
| -------------------- | ---------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| **Database**         | Non-superadmin must have `shop_id`; WorkOrders must contain `shop_id`. | ❌ `shop_id` is nullable; no role-based constraint; WorkOrder tables not present in provided SQL            |
| **UserService**      | Enforces visibility + action permission based on role + shop.          | ✅ Implemented via `canViewUser`, `canManageUser`, `checkFieldUpdatePermission`, and create-time shop rules |
| **ShopService**      | Admins may only modify their own shop.                                 | ❌ No RBAC; no actor passed into service/handler; any caller can list/update shops                          |
| **WorkOrderService** | All create/update actions validated against `(role, shop_id)`.         | ❌ Not implemented in backend code provided                                                                 |
|                      |

---

## 7. TL;DR Summary Table

| Area                          | SuperAdmin (spec) | Admin (spec)      | Backend Status Summary                                                  |
| ----------------------------- | ----------------- | ----------------- | ----------------------------------------------------------------------- |
| Needs `shop_id`?              | ❌                | ✔                 | ⚠️ Enforced on create in service, but not guaranteed at DB/update level |
| See all shops                 | ✔                 | ✔                 | ✅                                                                      |
| Edit shops                    | ✔ Any shop        | ✔ Own shop only   | ❌ No RBAC; any caller can edit any shop                                |
| See all admins                | ✔                 | ✔                 | ✅ Admins can see other admins, but cannot edit them                    |
| See other shops' staff        | ✔                 | ❌                | ✅ Admins only see staff in their own shop (plus all admins)            |
| Manage staff                  | ✔ Any shop        | ✔ Own shop only   | ✅ Admins can only manage staff in their own shop, via shop checks      |
| Create admin                  | ✔                 | ❌                | ✅ Enforced in `CreateUser`                                             |
| Create staff                  | ✔                 | ✔ (own shop only) | ✅ Enforced in `CreateUser` with shop scoping                           |
| Manage workorders (all shops) | ✔                 | ❌                | ❌ WorkOrder service not implemented yet                                |
| Manage workorders (own shop)  | ✔                 | ✔                 | ❌ WorkOrder service not implemented yet                                |

---

_This RBAC specification governs system-wide permission behavior and should be updated as backend code changes (especially when shop-based scoping and WorkOrder features are added)._
