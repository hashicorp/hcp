// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewGetReleaseV1Params creates a new GetReleaseV1Params object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetReleaseV1Params() *GetReleaseV1Params {
	return &GetReleaseV1Params{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetReleaseV1ParamsWithTimeout creates a new GetReleaseV1Params object
// with the ability to set a timeout on a request.
func NewGetReleaseV1ParamsWithTimeout(timeout time.Duration) *GetReleaseV1Params {
	return &GetReleaseV1Params{
		timeout: timeout,
	}
}

// NewGetReleaseV1ParamsWithContext creates a new GetReleaseV1Params object
// with the ability to set a context for a request.
func NewGetReleaseV1ParamsWithContext(ctx context.Context) *GetReleaseV1Params {
	return &GetReleaseV1Params{
		Context: ctx,
	}
}

// NewGetReleaseV1ParamsWithHTTPClient creates a new GetReleaseV1Params object
// with the ability to set a custom HTTPClient for a request.
func NewGetReleaseV1ParamsWithHTTPClient(client *http.Client) *GetReleaseV1Params {
	return &GetReleaseV1Params{
		HTTPClient: client,
	}
}

/*
GetReleaseV1Params contains all the parameters to send to the API endpoint

	for the get release v1 operation.

	Typically these are written to a http.Request.
*/
type GetReleaseV1Params struct {

	/* LicenseClass.

	     If specified, restrict responses to releases having this license
	class.  This is only used when fetching the "latest" release, as
	specifying a version to fetch necessarily also specifies the
	license class.

	*/
	LicenseClass *string

	/* Product.

	   The product name.
	*/
	Product string

	/* Version.

	   The product version or `latest`.
	*/
	Version string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get release v1 params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetReleaseV1Params) WithDefaults() *GetReleaseV1Params {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get release v1 params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetReleaseV1Params) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get release v1 params
func (o *GetReleaseV1Params) WithTimeout(timeout time.Duration) *GetReleaseV1Params {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get release v1 params
func (o *GetReleaseV1Params) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get release v1 params
func (o *GetReleaseV1Params) WithContext(ctx context.Context) *GetReleaseV1Params {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get release v1 params
func (o *GetReleaseV1Params) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get release v1 params
func (o *GetReleaseV1Params) WithHTTPClient(client *http.Client) *GetReleaseV1Params {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get release v1 params
func (o *GetReleaseV1Params) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithLicenseClass adds the licenseClass to the get release v1 params
func (o *GetReleaseV1Params) WithLicenseClass(licenseClass *string) *GetReleaseV1Params {
	o.SetLicenseClass(licenseClass)
	return o
}

// SetLicenseClass adds the licenseClass to the get release v1 params
func (o *GetReleaseV1Params) SetLicenseClass(licenseClass *string) {
	o.LicenseClass = licenseClass
}

// WithProduct adds the product to the get release v1 params
func (o *GetReleaseV1Params) WithProduct(product string) *GetReleaseV1Params {
	o.SetProduct(product)
	return o
}

// SetProduct adds the product to the get release v1 params
func (o *GetReleaseV1Params) SetProduct(product string) {
	o.Product = product
}

// WithVersion adds the version to the get release v1 params
func (o *GetReleaseV1Params) WithVersion(version string) *GetReleaseV1Params {
	o.SetVersion(version)
	return o
}

// SetVersion adds the version to the get release v1 params
func (o *GetReleaseV1Params) SetVersion(version string) {
	o.Version = version
}

// WriteToRequest writes these params to a swagger request
func (o *GetReleaseV1Params) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.LicenseClass != nil {

		// query param license_class
		var qrLicenseClass string

		if o.LicenseClass != nil {
			qrLicenseClass = *o.LicenseClass
		}
		qLicenseClass := qrLicenseClass
		if qLicenseClass != "" {

			if err := r.SetQueryParam("license_class", qLicenseClass); err != nil {
				return err
			}
		}
	}

	// path param product
	if err := r.SetPathParam("product", o.Product); err != nil {
		return err
	}

	// path param version
	if err := r.SetPathParam("version", o.Version); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
