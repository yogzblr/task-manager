Name:           automation-agent
Version:        1.0.0
Release:        1%{?dist}
Summary:        Automation Agent
License:        Proprietary
URL:            https://github.com/automation-platform/agent
Source0:        %{name}-%{version}.tar.gz

BuildArch:      x86_64
Requires:       systemd

%description
Cross-platform automation agent for the automation platform.

%prep
%setup -q

%build
# Build is done separately, binary is included in source

%install
mkdir -p %{buildroot}/usr/local/bin
mkdir -p %{buildroot}/etc/systemd/system
mkdir -p %{buildroot}/etc/sysconfig
mkdir -p %{buildroot}/var/lib/automation-agent
mkdir -p %{buildroot}/var/log/automation-agent

install -m 755 automation-agent %{buildroot}/usr/local/bin/automation-agent
install -m 644 deploy/linux/automation-agent.service %{buildroot}/etc/systemd/system/
install -m 644 deploy/linux/automation-agent.sysconfig %{buildroot}/etc/sysconfig/automation-agent

%pre
getent group automation-agent >/dev/null || groupadd -r automation-agent
getent passwd automation-agent >/dev/null || useradd -r -g automation-agent -d /var/lib/automation-agent -s /sbin/nologin automation-agent

%post
systemctl daemon-reload
systemctl enable automation-agent.service
systemctl start automation-agent.service

%preun
systemctl stop automation-agent.service || true
systemctl disable automation-agent.service || true

%files
/usr/local/bin/automation-agent
/etc/systemd/system/automation-agent.service
/etc/sysconfig/automation-agent
/var/lib/automation-agent
/var/log/automation-agent

%changelog
* Mon Jan 01 2024 Automation Platform Team <team@example.com> - 1.0.0-1
- Initial release
