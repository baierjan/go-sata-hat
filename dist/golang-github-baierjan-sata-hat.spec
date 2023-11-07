#
# spec file for package go-sata-hat
#
# Copyright (c) 2023 SUSE LLC
#
# All modifications and additions to the file contributed by third parties
# remain the property of their copyright owners, unless otherwise agreed
# upon. The license for this file, and modifications and additions to the
# file, is the same license as for the pristine package itself (unless the
# license for the pristine package is not an Open Source License, in which
# case the license is the MIT License). An "Open Source License" is a
# license that conforms to the Open Source Definition (Version 1.9)
# published by the Open Source Initiative.

# Please submit bugfixes or comments via https://bugs.opensuse.org/
#


Name:           golang-github-baierjan-sata-hat
Version:        0.2.0
Release:        0
Summary:        Quad SATA HAT Controller
License:        MIT
URL:            https://github.com/baierjan/go-sata-hat
Source:         %{name}-%{version}.tar.xz
Source1:        vendor.tar.gz
BuildRequires:  golang-packaging
BuildRequires:  systemd-rpm-macros
%{go_nostrip}

%description
A simple control software for Quad SATA HAT with fan and LCD diplay for Raspberry Pi 4.

%prep
%autosetup -p1 -a1

%build
go build \
   -mod=vendor \
   -buildmode=pie \
   ./src/fan-control

go build \
   -mod=vendor \
   -buildmode=pie \
   ./src/oled

%install
install -D -m0755 %{_builddir}/%{name}-%{version}/fan-control %{buildroot}%{_bindir}/fan-control
install -D -m0755 %{_builddir}/%{name}-%{version}/oled %{buildroot}%{_bindir}/sys-oled
install -D -m0644 dist/fan-control.service %{buildroot}%{_unitdir}/fan-control.service
install -D -m0644 dist/sys-oled.service %{buildroot}%{_unitdir}/sys-oled.service

%pre
%service_add_pre fan-control.service sys-oled.service

%post
%service_add_post fan-control.service sys-oled.service

%preun
%service_del_preun fan-control.service sys-oled.service

%postun
%service_del_postun fan-control.service sys-oled.service

%files
%license LICENSE
%doc README.md
%{_bindir}/fan-control
%{_bindir}/sys-oled
%{_unitdir}/fan-control.service
%{_unitdir}/sys-oled.service

%changelog
