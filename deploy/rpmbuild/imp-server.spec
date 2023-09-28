Name:	    imp-server
Version:    VERSION
Release:    0
Summary:    imp-server for devices

Group:	    Application/Server
License:    GPL
BuildRoot:  /root/rpmbuild/BUILDROOT/${name}-${VERSION}
Source:     imp-server-VERSION.tgz
Prefix:     /usr/local/imp-server
%description
    Device info collection and os install


%prep
%setup -q

%install
    mkdir -p %{buildroot}/usr/local/
    mkdir -p %{buildroot}/lib/systemd/system/
    cp -rp ./bin/  %{buildroot}/usr/local/imp/
    cp -rp ./conf/ %{buildroot}/usr/local/imp/
    cp deploy/systemd/imp-server.service %{buildroot}/lib/systemd/system/

%post
    systemctl daemon-reload
%preun
    if [ "$1" = "0" ] ; then
        systemctl stop imp-server
        systemctl disable imp-server
    fi

%postun
    if [ "$1" = "0" ] ; then
        rm -rf %{prefix}
        rm -f /lib/systemd/system/imp-server.service
        systemctl daemon-reload
    fi

%clean
    rm -rf %{buildroot}

%files
    %defattr(-,root,root,0755)
    %{prefix}
    /lib/systemd/system/imp-server.service

%changelog

