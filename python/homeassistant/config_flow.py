"""Config flow for the demeter integration."""

from __future__ import annotations

import logging
from typing import Any

import voluptuous as vol

from homeassistant.config_entries import ConfigFlow, ConfigFlowResult
from homeassistant.const import CONF_HOST
from homeassistant.core import HomeAssistant
from homeassistant.helpers.aiohttp_client import async_get_clientsession

from .api import DemeterClient, DemeterError
from .const import DOMAIN

_LOGGER = logging.getLogger(__name__)

STEP_USER_DATA_SCHEMA = vol.Schema(
    {
        vol.Required(CONF_HOST): str,
    }
)


def normalize_host(host: str) -> str:
    """Normalize the user-supplied host into a base URL.

    A scheme is added if omitted, so "demeter.local:8249" becomes
    "http://demeter.local:8249".
    """
    host = host.strip().rstrip("/")
    if "://" not in host:
        host = f"http://{host}"
    return host


async def validate_input(hass: HomeAssistant, data: dict[str, Any]) -> dict[str, Any]:
    """Validate the user input allows us to connect.

    Data has the keys from STEP_USER_DATA_SCHEMA with values provided by
    the user.
    """
    client = DemeterClient(async_get_clientsession(hass), data[CONF_HOST])

    try:
        await client.list_enclosures()
    except DemeterError as err:
        raise CannotConnect from err

    # Return info that you want to store in the config entry.
    return {"title": "Demeter"}


class DemeterConfigFlow(ConfigFlow, domain=DOMAIN):
    """Handle a config flow for demeter."""

    VERSION = 1

    async def async_step_user(
        self, user_input: dict[str, Any] | None = None
    ) -> ConfigFlowResult:
        """Handle the initial step."""
        errors: dict[str, str] = {}
        if user_input is not None:
            user_input = {
                **user_input,
                CONF_HOST: normalize_host(user_input[CONF_HOST]),
            }
            try:
                info = await validate_input(self.hass, user_input)
            except CannotConnect:
                errors["base"] = "cannot_connect"
            except Exception:
                _LOGGER.exception("Unexpected exception")
                errors["base"] = "unknown"
            else:
                await self.async_set_unique_id(user_input[CONF_HOST])
                self._abort_if_unique_id_configured()
                return self.async_create_entry(title=info["title"], data=user_input)

        return self.async_show_form(
            step_id="user", data_schema=STEP_USER_DATA_SCHEMA, errors=errors
        )


class CannotConnect(Exception):
    """Error to indicate we cannot connect."""
