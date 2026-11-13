package alor

import (
	"context"
	"fmt"
	"net/url"
)

// OrderGroupRequest — создание/редактирование группы заявок.
type OrderGroupRequest struct {
	Orders          []ObjOrderGroupItem `json:"orders"`
	ExecutionPolicy string              `json:"executionPolicy"`
}

/**
 * GetOrderGroups возвращает все группы заявок.
 */
func (c *Client) GetOrderGroups(ctx context.Context) (OrderGroupsInfo, error) {
	var out OrderGroupsInfo
	if err := c.GetJSON(ctx, "/commandapi/api/orderGroups", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

/**
 * CreateOrderGroup создаёт группу заявок.
 */
func (c *Client) CreateOrderGroup(ctx context.Context, req OrderGroupRequest) (*OrderGroupCreationSuccess, error) {
	var out OrderGroupCreationSuccess
	if err := c.PostJSON(ctx, "/commandapi/api/orderGroups", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * GetOrderGroup возвращает группу заявок по ID.
 */
func (c *Client) GetOrderGroup(ctx context.Context, orderGroupID string) (*OrderGroupInfo, error) {
	var out OrderGroupInfo
	path := fmt.Sprintf("/commandapi/api/orderGroups/%s", url.PathEscape(orderGroupID))
	if err := c.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * UpdateOrderGroup редактирует группу заявок.
 */
func (c *Client) UpdateOrderGroup(ctx context.Context, orderGroupID string, req OrderGroupRequest) error {
	path := fmt.Sprintf("/commandapi/api/orderGroups/%s", url.PathEscape(orderGroupID))
	return c.PutJSON(ctx, path, nil, req, nil)
}

/**
 * DeleteOrderGroup удаляет группу заявок.
 */
func (c *Client) DeleteOrderGroup(ctx context.Context, orderGroupID string) error {
	path := fmt.Sprintf("/commandapi/api/orderGroups/%s", url.PathEscape(orderGroupID))
	return c.DeleteJSON(ctx, path, nil, nil)
}
