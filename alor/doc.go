// Package alor — клиент Alor OpenAPI V2 (HTTP + WebSocket).
//
// Документация API: https://alor.dev/docs/
// Спецификация OpenAPI: api/openapi/WarpOpenAPIv2.yml
//
// Возможности:
//   - авторизация через refresh → JWT (oauth.alor.ru / oauthdev.alor.ru);
//   - полный набор HTTP-ручек market data, client info, orders, stop-orders, order groups;
//   - WebSocket /ws — рыночные и портфельные подписки с reconnect и resubscribe;
//   - WebSocket /cws — торговые команды (create/update/delete limit|market|stop|stopLimit).
//
// Контуры: Production (api.alor.ru) и Dev (apidev.alor.ru).
package alor
