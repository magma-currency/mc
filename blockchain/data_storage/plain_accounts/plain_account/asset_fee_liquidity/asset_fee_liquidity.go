package asset_fee_liquidity

import (
	"bytes"
	"errors"
	"pandora-pay/config/config_assets"
	"pandora-pay/config/config_coins"
	"pandora-pay/helpers"
)

type AssetFeeLiquidity struct {
	helpers.SerializableInterface `json:"-"`
	AssetId                       helpers.HexBytes `json:"assetId"`
	Rate                          uint64           `json:"rate"`
	LeadingZeros                  byte             `json:"leadingZeros"`
}

func (self *AssetFeeLiquidity) Validate() error {
	if len(self.AssetId) != config_coins.ASSET_LENGTH {
		return errors.New("AssetId length is invalid")
	}

	if bytes.Equal(self.AssetId, config_coins.NATIVE_ASSET_FULL) {
		return errors.New("AssetId NATIVE_ASSET_FULL is not allowed")
	}
	if self.LeadingZeros > config_assets.ASSETS_DECIMAL_SEPARATOR_MAX_BYTE {
		return errors.New("Invalid Leading Zeros")
	}

	return nil
}

func (self *AssetFeeLiquidity) Serialize(w *helpers.BufferWriter) {
	w.Write(self.AssetId)
	w.WriteUvarint(self.Rate)
	w.WriteByte(self.LeadingZeros)
}

func (self *AssetFeeLiquidity) Deserialize(r *helpers.BufferReader) (err error) {
	if self.AssetId, err = r.ReadBytes(config_coins.ASSET_LENGTH); err != nil {
		return
	}
	if self.Rate, err = r.ReadUvarint(); err != nil {
		return
	}

	if self.LeadingZeros, err = r.ReadByte(); err != nil {
		return
	}

	return
}